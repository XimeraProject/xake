package main

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/fatih/color"
	prefixed "github.com/kisonecat/logrus-prefixed-formatter"
	"github.com/urfave/cli"
	"net/url"
	"os"
	"sort"
)

var log = logrus.New()
var repository string
var keyFingerprint string
var ximeraUrl *url.URL

func init() {
	formatter := new(prefixed.TextFormatter)
	formatter.DisableTimestamp = true
	formatter.DisableUppercase = true
	log.Formatter = formatter
}

func main() {
	app := cli.NewApp()

	app.Name = "xake"
	app.Usage = "a build tool (make) for Ximera"
	app.Version = "0.2.1"
	app.EnableBashCompletion = true

	fmt.Printf("This is xake, Version " + app.Version + "\n\n")

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
		cli.BoolFlag{
			Name:  "no-color, C",
			Usage: "Disable color",
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
		cli.StringFlag{
			Name:  "key, k",
			Value: keyFingerprint,
			Usage: "Request authorization using GPG key with `FINGERPRINT`",
		},
		cli.StringFlag{
			Name:  "url, U",
			Value: "http://localhost:3000/",
			Usage: "Use the Ximera server hosted at `URL`",
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
			Name:    "name",
			Aliases: []string{"n"},
			Usage:   "provide a name for this repository",
			Action: func(c *cli.Context) error {
				name := c.Args().Get(0)
				Name(name)
				return nil
			},
		},
		{
			Name:    "bake",
			Aliases: []string{"b"},
			Usage:   "compile all the files in the repository",
			Action: func(c *cli.Context) error {
				workers := c.Int("jobs")
				if workers == 0 {
					workers = 2
				}
				return Bake(workers)
			},
		},
		{
			Name:    "frost",
			Aliases: []string{"f, ice"},
			Usage:   "add a publication tag to the repository",
			Action: func(c *cli.Context) error {
				err := Frost()
				if err != nil {
					log.Error(err)
				}
				return err

			},
		},

		{
			Name:    "serve",
			Aliases: []string{"s"},
			Usage:   "push the publication tag to the server",
			Action: func(c *cli.Context) error {
				err := Serve()
				if err != nil {
					log.Error(err)
				}
				return err
			},
		},

		{
			Name:    "view",
			Aliases: []string{"v"},
			Usage:   "view a picture of a piece of a cake",
			Action: func(c *cli.Context) error {
				EasterEgg()
				return nil
			},
		},

		{
			Name:    "information",
			Aliases: []string{"i", "info"},
			Usage:   "display information about the repository",
			Action: func(c *cli.Context) error {
				files, _, err := NeedingCompilation(repository)
				if err != nil {
					log.Error(err)
					return err
				}
				for _, file := range files {
					log.Warn(fmt.Sprintf("%s needs to be compiled", file))
				}
				return nil
			},
		},
	}

	app.Before = func(c *cli.Context) error {
		if c.Bool("verbose") {
			log.Level = logrus.DebugLevel
		}

		if c.Bool("no-color") {
			color.NoColor = true
			plainLogs := new(prefixed.TextFormatter)
			plainLogs.DisableColors = true
			plainLogs.DisableTimestamp = true
			log.Formatter = plainLogs
		}

		repository = c.String("repository")
		repository, err := FindRepositoryAmongParentDirectories(repository)
		if err != nil {
			return err
		}
		log.Debug("Using repository " + repository)

		keyFingerprint = c.String("key")
		// Failing to be able to resolve the key is not a fatal error,
		// because you don't necessarily need to have GPG installed in
		// order to make use of xake
		keyFingerprint, _ = ResolveKeyToFingerprint(keyFingerprint)
		log.Debug("Using GPG key " + keyFingerprint)

		urlString := c.String("url")
		// This should actually default to whatever the ximera remote is in the current repo
		ximeraUrl, err = url.Parse(urlString)
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
		//files, err := NeedingCompilation(repository)

		//log.Error(err)
		//log.Error(files)

		return nil
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	app.Run(os.Args)
}
