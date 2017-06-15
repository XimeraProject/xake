package main

import (
	//	"encoding/json"
	"errors"
	"fmt"
	"github.com/libgit2/git2go"
	//	"io/ioutil"
	"github.com/cavaliercoder/grab"
	"gopkg.in/cheggaaa/pb.v1"
	"net/url"
	//	"path/filepath"
	//	"regexp"
	"os"
	"strings"
	"time"
)

func DownloadData() error {
	log.Debug("Opening repository " + repository)
	repo, err := git.OpenRepository(repository)
	if err != nil {
		return err
	}

	remote, err := repo.Remotes.Lookup("ximera")
	if err != nil {
		return err
	}

	u, err := url.Parse(remote.Url())
	if err != nil {
		return err
	}

	u.Path = u.Path + "/log"

	config, err := repo.Config()
	if err != nil {
		return err
	}

	extraHeader, err := config.LookupString("http." + ximeraUrl.String() + ".extraHeader")
	if err != nil {
		return err
	}

	splitHeader := strings.SplitN(extraHeader, ":", 2)
	if len(splitHeader) != 2 {
		return errors.New("No authorization token found in .git config")
	}

	// create client
	client := grab.NewClient()
	req, _ := grab.NewRequest(".", u.String())

	req.HTTPRequest.Header.Set("Authorization", splitHeader[1])

	// start download
	fmt.Printf("Downloading records from %v...\n", req.URL())
	resp := client.Do(req)
	if resp.HTTPResponse.StatusCode == 500 {
		return errors.New("Error downloading log.")
	}

	// start UI loop
	t := time.NewTicker(100 * time.Millisecond)
	defer t.Stop()

	var bar *pb.ProgressBar
	bar = pb.StartNew(int(resp.Size))
	bar.ShowTimeLeft = true
	bar.SetUnits(pb.U_BYTES)
	bar.Start()

Loop:
	for {
		select {
		case <-t.C:
			bar.Set(int(resp.BytesComplete()))
		case <-resp.Done:
			// download is complete
			break Loop
		}
	}

	// check for errors
	if err := resp.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Download failed: %v\n", err)
		os.Exit(1)
	}

	bar.Set(int(resp.BytesComplete()))
	bar.FinishPrint("Downloaded data.")

	//fmt.Printf("Download saved to ./%v \n", resp.Filename)

	return nil
}
