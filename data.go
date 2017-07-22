package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/libgit2/git2go"
	//	"io/ioutil"
	"github.com/cavaliercoder/grab"
	"gopkg.in/cheggaaa/pb.v1"
	"net/url"
	//	"path/filepath"
	"github.com/funny/binary"
	//	"regexp"
	"github.com/golang/snappy"
	"os"
	"strings"
	"time"
)

func DownloadData() error {
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

	u.Path = u.Path + "/log.sz"

	config, err := repo.Config()
	if err != nil {
		return err
	}

	// BADBAD: should normalize ximeraUrl here

	extraHeader, err := config.LookupString("http." + ximeraUrl.String() + ".extraHeader")
	if err != nil {
		return err
	}

	splitHeader := strings.SplitN(extraHeader, ":", 2)
	if len(splitHeader) != 2 {
		return errors.New("No authorization token found in .git config")
	}

	// create client
	client := grab.NewClient()
	req, _ := grab.NewRequest(".", u.String())

	req.HTTPRequest.Header.Set("Authorization", splitHeader[1])

	// start download
	fmt.Printf("Downloading records from %v...\n", req.URL())
	resp := client.Do(req)
	if resp.HTTPResponse.StatusCode == 500 {
		return errors.New("Error downloading log.")
	}

	// start UI loop
	t := time.NewTicker(100 * time.Millisecond)
	defer t.Stop()

	var bar *pb.ProgressBar
	bar = pb.StartNew(int(resp.Size))
	bar.ShowTimeLeft = true
	bar.SetUnits(pb.U_BYTES)
	bar.Start()

Loop:
	for {
		select {
		case <-t.C:
			bar.Set(int(resp.BytesComplete()))
		case <-resp.Done:
			// download is complete
			break Loop
		}
	}

	// check for errors
	if err := resp.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Download failed: %v\n", err)
		os.Exit(1)
	}

	bar.Set(int(resp.BytesComplete()))
	bar.FinishPrint("Downloaded data.")

	fmt.Printf("Download saved to ./%v \n", resp.Filename)

	return nil
}

func processEvents(process func(string) error) error {
	// The underlying assumption is that each block in the framed
	// snappy file corresponds to a single event.

	f, err := os.Open("log.sz")
	if err != nil {
		return err
	}
	defer f.Close()

	reader := binary.NewReader(bufio.NewReader(f))

	for {
		b, err := reader.ReadByte()
		if err != nil {
			return nil
		}
		length := reader.ReadUint24LE()

		// compressed blocks have type 0
		if b == 0 {
			_ = reader.ReadUint32LE()
			data := reader.ReadBytes(int(length - 4))
			payload, err := snappy.Decode(nil, data)
			if err != nil {
				return err
			}
			process(string(payload))
		} else {
			// Skip all other blocks
			_ = reader.ReadBytes(int(length))
		}
	}

	return nil
}

func DumpEventsAsJSON() error {
	dump := func(payload string) error {
		fmt.Printf("%s,\n", string(payload))
		return nil
	}

	fmt.Printf("[\n")
	processEvents(dump)
	fmt.Printf("]\n")
	return nil
}

func DumpEventsAsCSV() error {
	dump := func(payload string) error {
		var event map[string]interface{}
		if err := json.Unmarshal([]byte(payload), &event); err != nil {
			return err
		}

		verb := event["verb"].(map[string]interface{})
		verbDisplay := verb["display"].(map[string]interface{})
		verbDisplayEnglish := verbDisplay["en-US"].(string)

		fmt.Printf("%s,%s,%s\n", event["actor"].(string), verbDisplayEnglish, event["timestamp"].(string))

		return nil
	}

	fmt.Printf("actor,verb,timestamp\n")

	processEvents(dump)
	return nil
}
