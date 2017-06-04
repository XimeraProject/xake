package main

import (
	"bufio"
	"io"
	"os/exec"
	"regexp"
	"strings"
)

func firstKey() (string, error) {
	cmdName := "gpg"
	cmdArgs := []string{"--list-secret-keys", "--with-colons"}

	cmd := exec.Command(cmdName, cmdArgs...)
	cmdOut, err := cmd.Output()

	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(cmdOut)))
	for scanner.Scan() {
		line := scanner.Text()
		data := strings.Split(line, ":")
		if data[0] == "sec" {
			return data[4], nil
		}
	}

	return "", err
}

func defaultKey() (string, error) {
	cmdName := "gpgconf"
	cmdArgs := []string{"--list-options", "gpg"}

	cmd := exec.Command(cmdName, cmdArgs...)
	cmdOut, err := cmd.Output()

	if err != nil {
		return firstKey()
	}

	scanner := bufio.NewScanner(strings.NewReader(string(cmdOut)))
	for scanner.Scan() {
		line := scanner.Text()
		data := strings.Split(line, ":")
		if data[0] == "default-key" {
			re := regexp.MustCompile("[^A-Fa-f0-9]")
			return re.ReplaceAllString(data[len(data)-1], ""), nil
		}
	}

	return string(cmdOut), err
}

func normalizeKey(keyId string) (string, error) {
	cmdName := "gpg"
	cmdArgs := []string{"--with-colons", "--fingerprint", keyId}

	cmd := exec.Command(cmdName, cmdArgs...)
	cmdOut, err := cmd.Output()

	if err != nil {
		return keyId, err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(cmdOut)))
	for scanner.Scan() {
		line := scanner.Text()
		data := strings.Split(line, ":")
		if data[0] == "fpr" {
			return data[len(data)-2], nil
		}
	}

	return keyId, nil
}

func ResolveKeyToFingerprint(keyId string) (keyFingerprint string, err error) {
	if len(keyId) != 0 {
		keyFingerprint = keyId
	} else {
		keyFingerprint, err = defaultKey()
	}

	keyFingerprint, err = normalizeKey(keyFingerprint)
	return
}

func Decrypt(src io.Reader) (string, error) {
	cmdName := "gpg"
	cmdArgs := []string{"--decrypt"}
	cmd := exec.Command(cmdName, cmdArgs...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}

	go func() {
		defer stdin.Close()
		io.Copy(stdin, src)
	}()

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(out), nil
}
