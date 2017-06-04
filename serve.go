package main

import (
	"github.com/fatih/color"
	"github.com/libgit2/git2go"
	"os"
	"os/exec"
	"strings"
)

func gitPushXimera(ref string) error {
	args := []string{"push", "ximera"}

	if len(ref) > 0 {
		args = []string{"push", "ximera", ref}
	}

	if ref == "master" {
		args = []string{"push", "--set-upstream", "ximera", ref}
	}

	yellow := color.New(color.FgYellow)
	yellow.Println("git " + strings.Join(args, " "))

	command := exec.Command("git", args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Start()
	return command.Wait()
}

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
		log.Error("Did you forget to perform a `xake frost` ?")
		return err
	}

	_, err = repo.Remotes.Lookup("ximera")
	if err != nil {
		log.Error("I could not find a ximera remote.")
		log.Error("Did you forget to perform a `xake name` ?")
		return err
	}

	err = gitPushXimera("")
	if err != nil {
		return err
	}

	err = gitPushXimera(tagName)
	if err != nil {
		return nil
	}

	// It would be better if it printed the URL here
	log.Info("The xake is served.")

	return nil
}
