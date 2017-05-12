package main

import (
	"fmt"
	"github.com/briandowns/spinner"
	"time"
)

func Name(name string) error {
	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	s.Prefix = fmt.Sprintf("Getting bearer token for '%s'... ", name)
	s.Start()

	token, err := GetRepositoryToken(name)
	if err != nil {
		log.Error(err)
		return err
	}
	s.Stop()
	fmt.Printf("Received bearer token for '%s'\n", name)

	fmt.Printf("Token is '%s'\n", token)

	// BADBAD: the token should now be loaded into the local gitconfig

	return nil
}
