package main

import (
	"fmt"
	"github.com/libgit2/git2go"
	"os"
	"path/filepath"
)

func exists(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}

func choose(ss []string, test func(string) bool) (ret []string) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}

func test(repo *git.Repository) error {
	headReference, err := repo.Head()
	if err != nil {
		return err
	}

	ref, _ := repo.References.Lookup(headReference.Name())
	fmt.Println(headReference.Name())
	fmt.Println(headReference.Target())
	fmt.Println(ref.Target())
	fmt.Println(ref.SymbolicTarget())

	return nil
}

func Publish() error {

	filenames, _ := NeedingPublication(repository)
	filenames = choose(filenames, exists)

	fmt.Println(filenames)

	log.Debug("Opening repository " + repository)
	repo, err := git.OpenRepository(repository)
	if err != nil {
		return err
	}

	log.Debug("Opening index...")
	index, err := repo.Index()
	if err != nil {
		return err
	}

	for _, filename := range filenames {
		relativePath, err := filepath.Rel(repository, filename)
		if err != nil {
			return err
		}

		log.Debug("git add " + filename)
		err = index.AddByPath(relativePath)
		if err != nil {
			return err
		}
	}

	log.Debug("Writing tree...")
	oid, err := index.WriteTree()
	if err != nil {
		return err
	}

	treeObject, err := repo.Lookup(oid)
	if err != nil {
		return err
	}

	tree, err := treeObject.AsTree()
	if err != nil {
		return err
	}

	committer, err := repo.DefaultSignature()
	if err != nil {
		return err
	}
	author := committer
	message := "xake publish"

	headReference, err := repo.Head()
	if err != nil {
		return err
	}
	headCommit, err := repo.LookupCommit(headReference.Target())
	if err != nil {
		return err
	}
	sourceOid := headReference.Target()

	commitOid, err := repo.CreateCommit("HEAD", author, committer, message, tree, headCommit)

	headReference, err = repo.References.Lookup("HEAD")
	if err != nil {
		return err
	}
	if headReference.Type() == git.ReferenceSymbolic {
		branchReference, err := repo.References.Lookup(headReference.SymbolicTarget())
		if err != nil {
			return err
		}

		branchReference.SetTarget(sourceOid, "xake publish reverting back to source code")
	}

	// Create tag
	tagName := "refs/tags/publications/" + (sourceOid.String())
	tagReference, err := repo.References.Lookup(tagName)
	if err == nil {
		tagReference.SetTarget(commitOid, "xake re-publish")
	} else {
		repo.References.Create(tagName, commitOid, false, "xake publish")
	}

	fmt.Println("committed")
	fmt.Println(commitOid)
	fmt.Println((sourceOid.String()))

	return nil
}
