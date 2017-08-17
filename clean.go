package main

import (
	"gopkg.in/cheggaaa/pb.v1"
	"os"
	"path/filepath"
)

var deletableExtensions = []string{
	".aux",
	".4ct",
	".4tc",
	".oc",
	".md5",
	".dpth",
	".out",
	".jax",
	".idv",
	".lg",
	".tmp",
	".xref",
	".log",
	".auxlock",
	".dvi",
	".pdf",
	".html",
	".dpth",
	".ids",
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func isDeletable(path string) bool {
	return stringInSlice(filepath.Ext(path), deletableExtensions)
}

func RemoveBuiltFiles() error {

	filenames, err := TexFilesInRepository(repository)

	if err != nil {
		return err
	}

	included := make(map[string]bool)

	for _, filename := range filenames {
		images, err := IncludedImages(filename)
		if err == nil {
			for _, image := range images {
				included[image] = true
			}
		}
	}

	var toDelete []string

	var visit = func(path string, f os.FileInfo, err error) error {
		if isDeletable(path) {
			if !included[path] {
				toDelete = append(toDelete, path)
			}
		}

		return nil
	}

	err = filepath.Walk(repository, visit)
	if err != nil {
		return err
	}

	var bar *pb.ProgressBar
	bar = pb.StartNew(len(toDelete))
	bar.ShowTimeLeft = true
	bar.Start()

	for _, filename := range toDelete {
		err := os.Remove(filename)
		if err != nil {
			return err
		}
		bar.Increment()
	}

	bar.FinishPrint("Cleaned your working tree.")

	return nil
}
