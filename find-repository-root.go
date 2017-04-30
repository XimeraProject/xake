package main

import (
	"fmt"
	"github.com/libgit2/git2go"
	"path/filepath"
)

// isGitRepository returns true or false, as to whether or not we can
// open the given directory as a git repository
func isGitRepository(directory string) bool {
	_, err := git.OpenRepository(directory)

	if err != nil {
		return false
	}

	return true
}

func FindRepositoryAmongParentDirectories(directory string) (string, error) {

	for {
		log.Debug("Checking if " + directory + " is a git repository.")
		if isGitRepository(directory) {
			return directory, nil
		} else {
			log.Debug(directory + " is not a git repository.")
		}

		parent := filepath.Clean(filepath.Join(directory, ".."))
		if parent == directory {
			return "", fmt.Errorf("Could not find git repository in any parent directory.")
		}

		directory = parent
	}
}
