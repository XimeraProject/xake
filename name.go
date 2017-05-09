package main

import (
	"fmt"
	"github.com/catalinc/hashcash"
	"os/exec"
	"strings"
)

func proofOfWork(name string) (string, error) {
	var bits uint = 20

	// the hashcash command is probably faster than the pure golang implementation
	cmd := exec.Command("hashcash", []string{"-b", fmt.Sprintf("%d", bits), "-qm", name}...)
	cmdOut, err := cmd.Output()

	if err != nil {
		h := hashcash.New(bits, 8, "")
		return h.Mint(name)
	}

	return strings.TrimSpace(string(cmdOut)), nil
}

func Name(name string) error {

	fmt.Printf("Claiming %s\n", name)

	stamp, _ := proofOfWork(name)
	fmt.Println(stamp)

	return nil
}
