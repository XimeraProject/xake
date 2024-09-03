package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/catalinc/hashcash"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
)

var apiToken string

func proofOfWork(name string) (string, error) {
	var bits uint = 20

	// the hashcash command is (apparently) faster than the pure golang implementation
	cmd := exec.Command("hashcash", []string{"-b", fmt.Sprintf("%d", bits), "-qm", name}...)
	cmdOut, err := cmd.Output()

	if err != nil {
		h := hashcash.New(bits, 8, "")
		return h.Mint(name)
	}

	return strings.TrimSpace(string(cmdOut)), nil
}

func endpoint(path string) string {
	u, err := url.Parse(path)
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		log.Fatal(err)
	}

	return ximeraUrl.ResolveReference(u).String()
}

func requestToken(keyId string) (string, error) {
	log.Debug("Using keyid " + keyId)
	url := endpoint("gpg/token/" + keyId)
	log.Debug("Sending to " + url)
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		bodyBytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return "", err
		}
		bodyString := string(bodyBytes)
		return "", errors.New(bodyString)
	}

	challenge, err := Decrypt(response.Body)
	if err != nil {
		return "", err
		return "", errors.New("Could not decrypt the challenge at " + url) 
	}

	return challenge, nil
}

func RequestLtiSecret(keyId string, ltiKey string) (string, error) {
	log.Debug("Using GPG key with fingerprint " + keyId)
	url := endpoint("gpg/secret/" + ltiKey + "/" + keyId)
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		bodyBytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return "", err
		}
		bodyString := string(bodyBytes)
		return "", errors.New(bodyString)
	}

	challenge, err := Decrypt(response.Body)
	if err != nil {
		return "", errors.New("Could not decrypt the secret at " + url)
	}

	return challenge, nil
}

func saveToken() (err error) {
	log.Debug("Saving token using fingerprint " + keyFingerprint)
	apiToken, err = requestToken(keyFingerprint)
	return
}

func authorize(req *http.Request) error {
	if apiToken == "" {
		err := saveToken()
		if err != nil {
			return err
		}
	}
	header := "Bearer " + apiToken
	log.Debug("Authorization: " + header)
	req.Header.Add("Authorization", header)
	return nil
}

func request(verb string, path string) (resp *http.Response, err error) {
	client := &http.Client{}

	req, err := http.NewRequest(verb, endpoint(path), nil)
	if err != nil {
		return nil, err
	}

	err = authorize(req)
	if err != nil {
		return nil, err
	}

	u, _ := url.Parse(path)
	stamp, _ := proofOfWork(u.Path)
	if len(stamp) > 0 {
		req.Header.Add("X-Hashcash", stamp)
	}

	return client.Do(req)
}

func get(url string) (resp *http.Response, err error) {
	return request("GET", url)
}

func post(url string) (resp *http.Response, err error) {
	return request("POST", url)
}

func GetRepositoryUrl(repositoryName string) string {
	return endpoint(repositoryName + ".git")
}

func GetRepositoryToken(repositoryName string) (string, error) {
	response, err := post(repositoryName + ".git")
	if err != nil {
		return "", err
	}

	if response.StatusCode != 200 {
		bodyBytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return "", err
		}
		bodyString := string(bodyBytes)
		return "", errors.New(bodyString)
	}

	dec := json.NewDecoder(response.Body)
	for {
		var v map[string]interface{}
		if err := dec.Decode(&v); err != nil {
			return "", err
		} else {
			if token, ok := v["token"]; ok {
				if str, ok := token.(string); ok {
					return str, nil
				}
			}
		}
	}

	return "", err
}
