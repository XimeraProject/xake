package main

import (
	"crypto/sha1"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var extensions = []string{
	"aux",
	"4ct",
	"4tc",
	"oc",
	"md5",
	"dpth",
	"out",
	"jax",
	"idv",
	"lg",
	"tmp",
	"xref",
	"log",
	"auxlock",
	"dvi",
	"pdf",
	"scmd",
	"sout",
}

func clean(filename string) {
	for _, extension := range extensions {
		f := strings.TrimSuffix(filename, filepath.Ext(filename)) + "." + extension
		os.Remove(f)
	}
}

func pdflatex(filename string) ([]byte, error) {
	cmdName := "pdflatex"
	tikzexport := "\"\\PassOptionsToClass{tikzexport}{ximera}\\PassOptionsToClass{xake}{ximera}\\PassOptionsToClass{xake}{xourse}\\nonstopmode\\input{" + filepath.Base(filename) + "}\""
	cmdArgs := []string{"-file-line-error", "-shell-escape", tikzexport}

	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Dir = filepath.Dir(filename)

	cmdOut, err := cmd.Output()

	return cmdOut, err
}

func htlatex(filename string) ([]byte, error) {
	cmdName := "htlatex"
	cmdArgs := []string{filepath.Base(filename), "ximera,charset=utf-8,-css", " -cunihtf -utf8", "", "--interaction=nonstopmode -shell-escape -file-line-error"}

	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Dir = filepath.Dir(filename)

	cmdOut, err := cmd.Output()

	return cmdOut, err
}

func sage(filename string) ([]byte, error) {
	cmdName := "sage"
	cmdArgs := []string{filepath.Base(filename)}
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Dir = filepath.Dir(filename)

	cmdOut, err := cmd.Output()

	return cmdOut, err
}

func FindLabelAnchorsInHtml(htmlFilename string) ([]string, error) {
	var ids []string

	f, err := os.Open(htmlFilename)
	defer f.Close()
	if err != nil {
		return ids, err
	}

	doc, err := goquery.NewDocumentFromReader(f)
	if err != nil {
		return ids, err
	}

	doc.Find("a.ximera-label").Each(func(i int, s *goquery.Selection) {
		id, exists := s.Attr("id")
		if exists {
			ids = append(ids, id)
		}
	})

	return ids, err
}

func readTitleAndAbstract(htmlFilename string) (title string, abstract string, err error) {
	f, err := os.Open(htmlFilename)
	defer f.Close()
	if err != nil {
		return
	}

	doc, err := goquery.NewDocumentFromReader(f)
	if err != nil {
		return
	}

	title = doc.Find("title").Contents().Text()

	doc.Find("div.abstract").Each(func(i int, s *goquery.Selection) {
		var html string
		html, err = s.Html()
		if err == nil {
			abstract = html
		}
	})

	doc.Find("div.abstract p").Each(func(i int, s *goquery.Selection) {
		var html string
		html, err = s.Html()
		if err == nil {
			abstract = html
		}
	})

	return
}

func transformXourse(directory string, filename string, doc *goquery.Document) {
	log.Debug("Transforming xourse file " + filename)

	log.Debug("Remove the anchor links that htlatex is inserting")
	doc.Find("a").Each(func(_ int, s *goquery.Selection) {
		_, exists := s.Attr("id")
		if exists {
			s.Remove()
		}
	})

	log.Debug("Normalize the activity links")

	doc.Find("a.activity").Each(func(_ int, s *goquery.Selection) {
		href, exists := s.Attr("href")

		if exists {
			// BADBAD: do I need this?
			texActivityFilename := filepath.Clean(filepath.Join(filepath.Dir(filename), href))
			activityFilename := strings.TrimSuffix(texActivityFilename, ".tex") + ".html"

			// Unfortunately xourse files links are relative to repo root
			href, _ = filepath.Rel(directory, texActivityFilename)
			href = strings.TrimSuffix(href, ".tex")
			s.SetAttr("href", href)

			log.Debug("Reading title and abstract from " + activityFilename)
			title, abstract, err := readTitleAndAbstract(activityFilename)
			if err == nil {
				s.AppendHtml("<h2>" + title + "</h2><h3>" + abstract + "</h3>")
			}
		}
	})

	return
}

func transformHtml(directory string, filename string) error {
	htmlFilename := strings.TrimSuffix(filename, filepath.Ext(filename)) + ".html"

	f, err := os.Open(htmlFilename)
	defer f.Close()
	if err != nil {
		return err
	}

	doc, err := goquery.NewDocumentFromReader(f)
	if err != nil {
		return err
	}

	xourseFile := false
	doc.Find("meta[name=\"description\"]").Each(func(i int, s *goquery.Selection) {
		content, exists := s.Attr("content")

		if !exists {
			log.Warn(htmlFilename + " is missing a content attribute on its meta[name=\"description\"]")
		}

		if content == "xourse" {
			xourseFile = true
		}
	})

	log.Debug("Remove empty paragraphs of the form <p></p>")
	doc.Find("p:empty").Each(func(i int, s *goquery.Selection) {
		s.Remove()
	})

	log.Debug("Add <meta> tags for all dependencies")
	doc.Find("head").Each(func(_ int, s *goquery.Selection) {
		dependencies, err := LatexDependencies(filename)
		if err == nil {
			absoluteFilename, err := filepath.Abs(filepath.Join(directory, filename))
			if err == nil {
				dependencies = append(dependencies, absoluteFilename)
			}

			absoluteRoot, err := filepath.Abs(directory)
			if err != nil {
				return
			}

			for _, dependency := range dependencies {
				dependency, err := filepath.Rel(absoluteRoot, dependency)

				if err != nil {
					continue
				}

				f, err := os.Open(dependency)
				defer f.Close()

				if err != nil {
					continue
				}

				h := sha1.New()
				if _, err := io.Copy(h, f); err == nil {
					hash := fmt.Sprintf("%x", h.Sum(nil))
					s.AppendHtml("<meta name=\"dependency\" content=\"" +
						hash + " " +
						dependency + "\">")
				}
			}
		}
	})

	if xourseFile {
		transformXourse(directory, filename, doc)
	}

	html, err := doc.Html()
	if err != nil {
		return err
	}

	fileInfo, err := os.Stat(htmlFilename)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(htmlFilename, []byte(html), fileInfo.Mode())
	if err != nil {
		return err
	}

	return nil
}

func Compile(directory string, filename string) ([]byte, error) {

	log.Debug("Cleaning files associated with " + filename)
	clean(filename)

	log.Debug("Running pdflatex for " + filename)
	output, err := pdflatex(filename)
	if err != nil {
		log.Error(string(output))
		return output, err
	}

	sagetexFilename := strings.TrimSuffix(filename, filepath.Ext(filename)) + ".sagetex.sage"
	if _, err := os.Stat(sagetexFilename); !os.IsNotExist(err) {
		log.Debug("Running sage for " + filename)
		sage(sagetexFilename)
	}

	log.Debug("Running pdflatex again for " + filename)
	output, err = pdflatex(filename)
	if err != nil {
		log.Error(string(output))
		return output, err
	}

	log.Debug("Running htlatex on " + filename)
	output, err = htlatex(filename)
	if err != nil {
		log.Error(string(output))
		return output, err
	}

	log.Debug("Applying HTML transformations for " + filename)
	err = transformHtml(directory, filename)
	if err != nil {
		return []byte{}, err
	}

	return []byte{}, nil
}
