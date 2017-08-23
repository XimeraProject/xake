package main

import (
	git "github.com/libgit2/git2go"
	"os"
	"os/exec"
)

func Defrost() error {
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
	tag, err := repo.References.Lookup(tagName)
	if err != nil {
		log.Error("There is no publication tag corresponding to the repository HEAD.")
		log.Error("There is `xake frost` to scrape off!")
		return err
	}
	return scrape( repo, tag, sourceOid );
}

func scrape( repo *git.Repository, tag *git.Reference, sourceOid *git.Oid ) error {
	sourceCommit,_ := repo.LookupCommit( sourceOid )
	parent := sourceCommit.Parent(0)
	repo.ResetToCommit( parent, git.ResetSoft, nil )
	tagName := tag.Name()
	tag.Delete()
	args := []string{"push", "ximera", "+:refs/tags/" + tagName, "+master"}
	command := exec.Command("git", args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Start()
	return command.Wait()
}

