package main

import (
	"github.com/fatih/color"
	"github.com/libgit2/git2go"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

func gitPushXimera(ref string) error {
	args := []string{"push", "ximera"}

	if len(ref) > 0 {
		args = []string{"push", "ximera", "+"+ref}
	}

	if ref == "master" {
		args = []string{"push", "--set-upstream", "ximera", "+"+ref}
	}

	yellow := color.New(color.FgYellow)
	yellow.Println("git " + strings.Join(args, " "))

	command := exec.Command("git", args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Start()
	return command.Wait()
}

func gitResetHard(ref string) error {
	args := []string{"reset", "--hard"}

	if len(ref) > 0 {
		args = []string{"reset", "--hard", ref}
	}

	yellow := color.New(color.FgYellow)
	yellow.Println("git " + strings.Join(args, " "))

	command := exec.Command("git", args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Start()
	return command.Wait()
}

func gitResetSoft(ref string) error {
	args := []string{"reset", "--soft"}

	if len(ref) > 0 {
		args = []string{"reset", "--soft", ref}
	}

	yellow := color.New(color.FgYellow)
	yellow.Println("git " + strings.Join(args, " "))

	command := exec.Command("git", args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Start()
	return command.Wait()
}

func gitResetHead() error {
	args := []string{"reset", "HEAD"}

	yellow := color.New(color.FgYellow)
	yellow.Println("git " + strings.Join(args, " "))

	command := exec.Command("git", args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Start()
	return command.Wait()
}

func gitFetchXimera() error {
	args := []string{"fetch", "--tags", "ximera"}

	yellow := color.New(color.FgYellow)
	yellow.Println("git " + strings.Join(args, " "))

	command := exec.Command("git", args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Start()
	return command.Wait()
}

func gitCheckout(ref string) error {
	args := []string{"checkout", "-f"}

	if len(ref) > 0 {
		args = []string{"checkout", "-f", ref}
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

	remote, err := repo.Remotes.Lookup("ximera")
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

	// Produce a URL pointing to all the repo's content
	u, err := url.Parse(remote.Url())
	if err != nil {
		return nil
	}
	u.User = nil
	u.Path = strings.TrimSuffix(u.Path, ".git")

	log.Info("The xake is served at ", u)

	return nil
}

func ServePull() error {
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

	_, err = repo.Remotes.Lookup("ximera")
	if err != nil {
		log.Error("I could not find a ximera remote.")
		log.Error("Did you forget to perform a `xake name` ?")
		return err
	}

	err = gitFetchXimera()
	if err != nil {
		return nil
	}

	tagName := "refs/tags/publications/" + (sourceOid.String())

	err = gitResetHard(tagName)
	if err != nil {
		return nil
	}

	err = gitResetSoft(sourceOid.String())
	if err != nil {
		return nil
	}

	err = gitResetHead()
	if err != nil {
		return nil
	}

	log.Info("Copied published content into your working tree.")

	return nil
}
