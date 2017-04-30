package main

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"os"
	"sort"
)

var log = logrus.New()
var repository string

func init() {
	log.Formatter = new(prefixed.TextFormatter)
}

func main() {
	app := cli.NewApp()

	app.Name = "xake"
	app.Usage = "a build tool (make) for Ximera"
	app.Version = "0.2.0"
	app.EnableBashCompletion = true

	cli.VersionFlag = cli.BoolFlag{
		Name:  "version, V",
		Usage: "print the version",
	}

	// BADBAD: This should produce nicer error outputs
	w := log.Writer()
	defer w.Close()
	log.WriterLevel(logrus.ErrorLevel)
	cli.ErrWriter = w

	repository, _ = os.Getwd()

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose, v, debug, d",
			Usage: "Display additional debugging information",
		},
		cli.IntFlag{
			Name:  "jobs, j",
			Value: 2,
			Usage: "The number of processes to run in parallel",
		},
		cli.StringFlag{
			Name:  "repository, r",
			Value: repository,
			Usage: "Perform operations on the repository at `PATH`",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "compile",
			Aliases: []string{"c"},
			Usage:   "compile a .tex file into an .html file",
			Action: func(c *cli.Context) error {
				filename := c.Args().Get(0)
				log.Info("Compiling " + filename + " in .")
				_, err := Compile(".", filename)
				if err != nil {
					log.Error(err)
				}
				return nil
			},
		},
		{
			Name:    "bake",
			Aliases: []string{"b"},
			Usage:   "compile all the files in the repository",
			Action: func(c *cli.Context) error {
				return nil
			},
		},

		{
			Name:    "information",
			Aliases: []string{"i"},
			Usage:   "display information about the repository",
			Action: func(c *cli.Context) error {
				fmt.Print("Repository = " + repository)
				return nil
			},
		},
	}

	app.Before = func(c *cli.Context) error {
		if c.Bool("verbose") {
			log.Level = logrus.DebugLevel
		}

		repository = c.String("repository")
		repository, err := FindRepositoryAmongParentDirectories(repository)
		if err != nil {
			return err
		}

		// BADBAD: should be spun out to a goroutine which panics?
		if !IsXimeraClassFileInstalled() {
			return fmt.Errorf("Could not find a copy of ximera.cls, but xake requires that you install the ximeraLatex package.")
		}

		/*
			err := CheckXimeraVersion()
			if err != nil {
				log.Error(err)
			}*/

		//dependencies, _ := LatexDependencies("sample.tex")
		//b, err := IsClean("/home/jim/ximeraSample", "/home/jim/ximeraSample/sample.tex")
		files, err := NeedingCompilation(repository)

		log.Error(err)
		log.Error(files)

		return nil
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	app.Run(os.Args)
}
