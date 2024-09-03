package main

import (
	"encoding/json"
	"github.com/libgit2/git2go/v34"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
)

// locateXimeraCls uses kpsewhich to locate a full pathname for the
// ximera.cls file.
func locateXimeraCls() (string, error) {
	cmdName := "kpsewhich"
	cmdArgs := []string{"ximera.cls"}

	if cmdOut, err := exec.Command(cmdName, cmdArgs...).Output(); err != nil {
		return "", err
	} else {
		ximeraCls := strings.Replace(
			strings.SplitAfterN(string(cmdOut), "\n", 1)[0],
			"\n", "", -1)

		log.Debug("Found ximera.cls at "+ ximeraCls)

		return ximeraCls, nil
	}
}

/* IsXimeraClassFileInstalled returns true or false as to whether or
not pdflatex (via kpsewhich) can find ximera.cls */
func IsXimeraClassFileInstalled() bool {
	_, err := locateXimeraCls()
	if err != nil {
		return false
	}

	return true
}

// FetchXimeraGithubSha downloads the commit sha for ximeraLatex's
// HEAD on GitHub
func fetchXimeraClsGithubSha() (string, error) {
	client := &http.Client{}

	url := "https://api.github.com/repos/XimeraProject/ximeraLatex/commits/master"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	for {
		var v map[string]interface{}
		if err := dec.Decode(&v); err != nil {
			return "", err
		} else {
			if sha, ok := v["sha"]; ok {
				if str, ok := sha.(string); ok {
					return str, nil
				}
			}
		}
	}
}

// fetchXimeraLocalSha gets the HEAD commit from wherever the
// ximera.cls file is stored
func fetchXimeraClsLocalSha(ximeraClsPath string) (string, error) {
	log.Debug("Checking local SHA for repository in "+ ximeraClsPath)

	repo, err := git.OpenRepository(ximeraClsPath)

	if err != nil {
		return "", err
	}

	head, err := repo.Head()

	if err != nil {
		return "", err
	}

	return head.Target().String(), nil
}

func CheckXimeraVersion() error {
	ximeraClsFilename, err := locateXimeraCls()

	if err != nil {
		log.Warn("You need to install https://github.com/ximeraProject/ximeraLatex into ~/texmf/tex/latex")
		return err
	}

	ximeraClsPath := filepath.Dir(ximeraClsFilename)

	localSha, err := fetchXimeraClsLocalSha(ximeraClsPath)
	if err != nil {
		return err
	}

	remoteSha, err := fetchXimeraClsGithubSha()
	if err != nil {
		log.Warn("Failed checking ximera sha: ")
		log.Warn("Error")
	} else {
		if remoteSha != localSha {
			log.Warn("The version of ximeraLatex on GitHub differs from the version you have installed.")
			log.Warn("Use (cd " + ximeraClsPath + " && git checkout master && git pull) to update.")
		}
	}

	return nil
}
