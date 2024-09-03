package main

import (
	"fmt"
	"github.com/libgit2/git2go/v34"
	"net/url"
	"os"
)

func Name(name string) error {
	fmt.Sprintf("Getting bearer token for '%s'... ", name)

	token, err := GetRepositoryToken(name)
	if err != nil {
		log.Error(err)
		os.Exit(1)
		return err
	}
	fmt.Printf("Received bearer token for '%s'\n", name)

	fmt.Printf("Token is '%s'\n", token)
	// BADBAD: the token should now be loaded into the local gitconfig

	log.Debug("Opening repository " + repository)
	repo, err := git.OpenRepository(repository)
	if err != nil {
		log.Error(err)
		os.Exit(1)
		return err
	}

	config, err := repo.Config()
	if err != nil {
		log.Error(err)
		os.Exit(1)
		return err
	}

	err = config.SetString("http."+ximeraUrl.String()+".extraHeader", "Authorization: Bearer "+token)
	if err != nil {
		log.Error(err)
		os.Exit(1)
		return err
	}

	// Unfortunately, old versions of git don't support the
	// extraHeader option, so for now let's also include the token as
	// the password for basic auth
	u, err := url.Parse(GetRepositoryUrl(name))
	if err != nil {
		return err
	}
	u.User = url.UserPassword("xake", token)

	_, err = repo.Remotes.Lookup("ximera")
	if err != nil {
		_, err = repo.Remotes.Create("ximera", u.String())
		if err != nil {
			return err
		}
	}
	err = repo.Remotes.SetUrl("ximera", u.String())

	return nil
}
