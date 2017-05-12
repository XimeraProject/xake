package main

import (
	"fmt"
	"github.com/libgit2/git2go"
)

func Serve() error {
	log.Debug("Opening repository " + repository)
	repo, err := git.OpenRepository(repository)
	if err != nil {
		return err
	}

	headReference, err := repo.Head()
	if err != nil {
		return err
	}
	sourceOid := headReference.Target()

	tagName := "refs/tags/publications/" + (sourceOid.String())
	_, err = repo.References.Lookup(tagName)
	if err != nil {
		log.Error("There is no publication tag corresponding to the repository HEAD.")
		log.Error("Did you forget to `xake frost` ?")
		return nil
	}

	fmt.Println("git push ximera")
	fmt.Println("git push ximera " + tagName)

	return nil
}
