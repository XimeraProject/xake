package main

import (
	"encoding/json"
	"fmt"
	"github.com/libgit2/git2go"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
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

type githubRepository struct {
	Owner      string `json:"owner"`
	Repository string `json:"repository"`
}

type metadata struct {
	XakeVersion string                       `json:"xakeVersion"`
	Labels      map[string]string            `json:"labels"`
	Github      *githubRepository            `json:"github"`
	Xourses     map[string]map[string]string `json:"xourses"`
}

func Frost(xakeVersion string) error {

	log.Debug("Find the \\label{}s in .html files")
	labels, err := FindLabelAnchorsInRepository(repository)
	if err != nil {
		return err
	}

	log.Debug("Find xourse metadata in .html files")
	xourses, err := FindXoursesInRepository(repository)
	if err != nil {
		return err
	}

	log.Debug("Determine what files need to be published.")
	filenames, _ := NeedingPublication(repository)
	filenames = choose(filenames, exists)

	log.Debug("Opening repository " + repository)
	repo, err := git.OpenRepository(repository)
	if err != nil {
		return err
	}

	log.Debug("Write metadata.json to the repository root")

	var github *githubRepository
	githubHttps, _ := regexp.Compile("^https://github.com/([^/]+)/(.*)\\.git$")
	githubSsh, _ := regexp.Compile("^git@github.com:([^/]+)/(.*)\\.git$")

	remotes, _ := repo.Remotes.List()
	for _, remoteName := range remotes {
		remote, err := repo.Remotes.Lookup(remoteName)
		if err == nil {
			url := remote.Url()
			matches := githubHttps.FindStringSubmatch(url)
			if len(matches) > 0 {
				github = &githubRepository{Owner: matches[1], Repository: matches[2]}
			}

			matches = githubSsh.FindStringSubmatch(url)
			if len(matches) > 0 {
				github = &githubRepository{Owner: matches[1], Repository: matches[2]}
			}
		}
	}

	m := metadata{XakeVersion: xakeVersion, Labels: labels, Github: github, Xourses: xourses}

	bytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(repository, "metadata.json"), bytes, 0644)
	if err != nil {
		return err
	}

	filenames = append(filenames, filepath.Join(repository, "metadata.json"))

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

	commitOid, err := repo.CreateCommit("", author, committer, message, tree, headCommit)

	// Create or update tag
	tagName := "refs/tags/publications/" + (sourceOid.String())
	tagReference, err := repo.References.Lookup(tagName)
	var created string
	if err == nil {
		taggedCommit, err2 := repo.LookupCommit(tagReference.Target())
		if err2 == nil && oid.Equal(taggedCommit.TreeId()) {
			fmt.Printf("No changes since last frost, so we'll just chill.\n")
			return nil
		}
		tagReference.SetTarget(commitOid, "xake re-publish")
		created = "Updated"
	} else {
		repo.References.Create(tagName, commitOid, false, "xake publish")
		created = "Created"
	}
	fmt.Printf("%s publication commit %s... for commit %s...\n", created, commitOid.String()[0:7], sourceOid.String()[0:7])
	fmt.Printf("Your next step is probably `xake serve`\n")
	return nil
}
