package main

import (
	"errors"
	"fmt"
	"github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
	"github.com/libgit2/git2go"
	"gopkg.in/sorcix/irc.v2"
	"net"
	"net/url"
	"regexp"
	"strconv"
)

type Payload struct {
	UserId  string `json:"userId"`
	Name    string `json:"name"`
	Message string `json:"message"`
}

func Chat() error {
	log.Debug("Opening repository " + repository)
	repo, err := git.OpenRepository(repository)
	if err != nil {
		return err
	}

	remote, err := repo.Remotes.Lookup("ximera")
	if err != nil {
		return err
	}

	u, err := url.Parse(remote.Url())
	if err != nil {
		return err
	}

	secure := false
	if u.Scheme == "https" {
		secure = true
	}

	port, err := strconv.Atoi(u.Port())
	if err != nil {
		if secure {
			port = 443
		} else {
			port = 80
		}
	}

	c, err := gosocketio.Dial(
		gosocketio.GetUrl(u.Hostname(), port, secure),
		transport.GetDefaultWebsocketTransport())

	if err != nil {
		return err
	}

	token, set := u.User.Password()
	if !set {
		return errors.New("No token set in ximera remote")
	}

	path := u.Path
	pathRegexp, _ := regexp.Compile("^/(.*)\\.git$")
	matches := pathRegexp.FindAllStringSubmatch(path, -1)
	if len(matches) == 0 {
		return errors.New("No repository in ximera remote")
	}

	repository := matches[0][1]

	incoming := make(chan Payload)
	outgoing := make(chan Payload)

	err = c.On(gosocketio.OnConnection, func(h *gosocketio.Channel) {
		fmt.Println("Proxying chat traffic from " + u.Hostname())

		type Credentials struct {
			Repository string `json:"repository"`
			Token      string `json:"token"`
		}

		h.Emit("xake", Credentials{Repository: repository, Token: token})

		go func() {
			for {
				h.Emit("xake-chat", <-outgoing)
			}
		}()
	})

	err = c.On("xake-chat", func(h *gosocketio.Channel, args Payload) {
		incoming <- args
	})

	if err != nil {
		return err
	}

	err = c.On(gosocketio.OnDisconnection, func(h *gosocketio.Channel) {
		fmt.Println("Disconnected")
	})

	if err != nil {
		return err
	}

	ircServer(incoming, outgoing)

	c.Close()

	return nil
}

func handleClient(conn net.Conn, incoming chan Payload, outgoing chan Payload) error {
	var nick string

	dec := irc.NewDecoder(conn)
	enc := irc.NewEncoder(conn)

	//(RPL_WELCOME), the server
	//name and version (RPL_YOURHOST), the server birth information
	//(RPL_CREATED), available user and channel modes (RPL_MYINFO),

	go func() {
		for {
			m := <-incoming

			enc.Encode(&irc.Message{
				Prefix:  &irc.Prefix{Name: m.UserId},
				Command: irc.PRIVMSG,
				Params:  []string{nick, m.Message}})
		}
	}()

	go func() {
		for {
			message, err := dec.Decode()
			if err != nil {
				return
			}

			if message.Command == "NICK" {
				nick = message.Params[0]

				enc.Encode(&irc.Message{
					Prefix:  &irc.Prefix{Host: "localhost"},
					Command: irc.RPL_WELCOME,
					Params:  []string{nick, "Welcome to Ximera"}})
				enc.Encode(&irc.Message{
					Prefix:  &irc.Prefix{Name: nick, User: nick, Host: "localhost"},
					Command: irc.RPL_MOTDSTART,
					Params:  []string{"-", "localhost", "Welcome to Ximera"}})
				enc.Encode(&irc.Message{
					Prefix:  &irc.Prefix{Name: nick, User: nick, Host: "localhost"},
					Command: irc.RPL_MOTD,
					Params:  []string{"-", "for proxying learner chats..."}})
				enc.Encode(&irc.Message{
					Prefix:  &irc.Prefix{Name: nick, User: nick, Host: "localhost"},
					Command: irc.RPL_ENDOFMOTD,
					Params:  []string{"-", "ready."}})

				continue
			}

			if message.Command == "PING" {
				enc.Encode(&irc.Message{
					Command: irc.PONG})
				continue
			}

			if message.Command == "PRIVMSG" {
				outgoing <- Payload{UserId: message.Params[0],
					Message: message.Params[1]}

				continue
			}

			fmt.Println(message)
		}
	}()

	return nil
}

func ircServer(incoming chan Payload, outgoing chan Payload) error {
	connections := make(chan net.Conn)

	ln, _ := net.Listen("tcp", ":6667")
	defer ln.Close()
	fmt.Println("Listening on localhost:6667")

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				continue
			}
			connections <- conn
		}
	}()

	for {
		conn := <-connections

		go handleClient(conn, incoming, outgoing)
	}

	return nil
}
