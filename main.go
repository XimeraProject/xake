package main

import (
	"fmt"
	"github.com/fatih/color"
	prefixed "github.com/kisonecat/logrus-prefixed-formatter"
	"github.com/sirupsen/logrus"
	//"github.com/tcnksm/go-latest"
	"github.com/urfave/cli"
	"net/url"
	"os"
	"sort"
	"sync"
)

var log = logrus.New()
var repository string
var keyFingerprint string
var ximeraUrl *url.URL
var workers int

func init() {
	formatter := new(prefixed.TextFormatter)
	formatter.DisableTimestamp = true
	formatter.DisableUppercase = true
	log.Formatter = formatter
}

func main() {
	var group sync.WaitGroup

	app := cli.NewApp()

	app.Name = "xake"
	app.Usage = "a build tool (make) for Ximera"
	app.Version = "0.8.30"

	// Check to see if this is the newest version Humorously,
	// go-latest depends on go>=1.7 because that was when "context"
	// was added to the main go libraries
	/*
		go func() {
			group.Add(1)
			githubTag := &latest.GithubTag{
				Owner:             "XimeraProject",
				Repository:        "xake",
				FixVersionStrFunc: latest.DeleteFrontV(),
			}
			res, err := latest.Check(githubTag, app.Version)
			if err != nil {
				log.Warn("Could not check if " + app.Version + " is the latest version.")
				log.Warn(err)
			} else {
				if res.Outdated {
					log.Error(app.Version + " is not the latest version of xake.")
					log.Error(fmt.Sprintf("You should upgrade to %s", res.Current))
				}
			}

			group.Done()
		}()
	*/

	app.EnableBashCompletion = true

	// Do not display all the aliases
	cli.AppHelpTemplate = `NAME:
   {{.Name}}{{if .Usage}} - {{.Usage}}{{end}}

USAGE:
   {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}{{if .Version}}{{if not .HideVersion}}

VERSION:
   {{.Version}}{{end}}{{end}}{{if .Description}}

DESCRIPTION:
   {{.Description}}{{end}}{{if len .Authors}}

AUTHOR{{with $length := len .Authors}}{{if ne 1 $length}}S{{end}}{{end}}:
   {{range $index, $author := .Authors}}{{if $index}}
   {{end}}{{$author}}{{end}}{{end}}{{if .VisibleCommands}}

COMMANDS:{{range .VisibleCategories}}{{if .Name}}
   {{.Name}}:{{end}}{{range .VisibleCommands}}
     {{.Name}}{{"\t"}}{{.Usage}}{{end}}{{end}}{{end}}{{if .VisibleFlags}}

GLOBAL OPTIONS:
   {{range $index, $option := .VisibleFlags}}{{if $index}}
   {{end}}{{$option}}{{end}}{{end}}{{if .Copyright}}

COPYRIGHT:
   {{.Copyright}}{{end}}
`

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
			Name:  "quiet, q",
			Usage: "Do not display extra information",
		},
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
			Value: "https://ximera.osu.edu/",
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
					log.Error("Could not compile " + filename)
					os.Exit(1)
				}
				return nil
			},
		},
		{
			Name:    "clean",
			Aliases: []string{"k"},
			Usage:   "remove built files from the working tree",
			Action: func(c *cli.Context) error {
				pathname := ""
				if len(c.Args()) > 0 {
					pathname = c.Args().Get(0)
				}

				err := RemoveBuiltFiles(pathname)
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
			Name:    "lti",
			Aliases: []string{"l"},
			Usage:   "fetch credentials for an LTI connection",
			Action: func(c *cli.Context) error {
				ltiKey := c.Args().Get(0)
				ltiSecret, err := RequestLtiSecret(keyFingerprint, ltiKey)

				if err != nil {
					log.Error(err)
					return err
				}

				fmt.Printf("LTI key: %s\n", ltiKey)
				fmt.Printf(" secret: %s\n", ltiSecret)

				return nil
			},
		},

		{
			Name: "bake",
			// I have so much trouble typing this word
			Aliases: []string{"b", "abke", "beak", "beka", "bkae", "bkea", "eabk", "eakb", "ebak", "ebka", "ekab", "ekba", "kabe", "kaeb", "kbae", "kbea", "keab", "keba"},
			Usage:   "compile all the files in the repository",
			Action: func(c *cli.Context) error {
				return Bake(workers)
			},
		},
		{
			Name:    "frost",
			Aliases: []string{"f, ice"},
			Usage:   "add a publication tag to the repository",
			Action: func(c *cli.Context) error {
				// BADBAD: should verify that we've commited the compiled source files
				err := DisplayErrorsAboutUncommittedTexFiles(repository)
				if err != nil {
					log.Error(err)
				} else {
					err = Frost(app.Version)
					if err != nil {
						log.Error(err)
					}
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
			Name:    "chat",
			Aliases: []string{"c"},
			Usage:   "run a chat server",
			Action: func(c *cli.Context) error {
				err := Chat()
				if err != nil {
					log.Error(err)
				}
				return err
			},
		},

		{
			Name:    "data",
			Aliases: []string{"t"},
			Usage:   "operate on the learning record store",
			Subcommands: []cli.Command{
				{
					Name:  "json",
					Usage: "dump log.sz as json entries",
					Action: func(c *cli.Context) error {
						err := DumpEventsAsJSON()
						if err != nil {
							log.Error(err)
						}
						return err
					},
				},
				{
					Name:  "csv",
					Usage: "dump log.sz as comma-separated values",
					Action: func(c *cli.Context) error {
						err := DumpEventsAsCSV()
						if err != nil {
							log.Error(err)
						}
						return err
					},
				},
				{
					Name:  "download",
					Usage: "download events to the log.sz file",
					Action: func(c *cli.Context) error {
						err := DownloadData()
						if err != nil {
							log.Error(err)
						}
						return err
					},
				},
			},
		},

		{
			Name:    "view",
			Hidden:  true,
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
		if !c.Bool("quiet") {
			fmt.Printf("This is xake, Version " + app.Version + "\n\n")
		}

		if c.Bool("verbose") {
			log.Level = logrus.DebugLevel
		}

		workers = c.Int("jobs")
		if workers == 0 {
			workers = 2
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
		// BADBAD: This should actually default to whatever the ximera remote is in the current repo
		ximeraUrl, err = url.Parse(urlString)
		if err != nil {
			return err
		}

		// Check to see if we can find ximeraLatex
		go func() {
			group.Add(1)
			if !IsXimeraClassFileInstalled() {
				log.Error("Could not find a copy of ximera.cls.")
				log.Error("Xake requires that you install the ximeraLatex package.")
			}
			group.Done()
		}()

		// Check to see if the version of ximeraLatex is good
		go func() {
			group.Add(1)
			err := CheckXimeraVersion()
			if err != nil {
				log.Error(err)
			}
			group.Done()
		}()

		//dependencies, _ := LatexDependencies("sample.tex")
		//b, err := IsClean("/home/jim/ximeraSample", "/home/jim/ximeraSample/sample.tex")
		//files, err := NeedingCompilation(repository)

		//log.Error(err)
		//log.Error(files)

		return nil
	}

	app.CommandNotFound = func(c *cli.Context, command string) {
		log.Error("I do not know how to '" + command + "'.")
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	app.Run(os.Args)

	group.Wait()
}
