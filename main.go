package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	createRelease "create/createRelease"

	"github.com/pivotal-cf/jhanda/flags"
	"github.com/pivotal-cf/winfs-injector/injector"
	"github.com/pivotal-cf/winfs-injector/winfsinjector"
)

const usageText = `winfs-injector
winfs-injector injects the Windows 2016 root file system into the Windows 2016 Runtime Tile.

Usage: winfs-injector
  --input-tile, -i   path to input tile (example: /path/to/input.pivotal)
  --output-tile, -o  path to output tile (example: /path/to/output.pivotal)
  --help, -h         prints this usage information
`

func main() {
	var arguments struct {
		InputTile  string `short:"i" long:"input-tile"  description:"path to input tile (example: /path/to/input.pivotal)"   default:""`
		OutputTile string `short:"o" long:"output-tile" description:"path to output tile (example: /path/to/output.pivotal)" default:""`
		Help       bool   `short:"h" long:"help"        description:"prints this usage information"                             default:"false"`
	}

	_, err := flags.Parse(&arguments, os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	if arguments.Help || arguments.InputTile == "" || arguments.OutputTile == "" {
		printUsage()
		return
	}

	var tileInjector = injector.NewTileInjector()
	var zipper = injector.NewZipper()
	var releaseCreator = createRelease.ReleaseCreator{}

	wd, err := ioutil.TempDir("", "")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(wd)

	app := winfsinjector.NewApplication(releaseCreator, tileInjector, zipper)
	err = app.Run(arguments.InputTile, arguments.OutputTile, wd)
	if err != nil {
		log.Fatal(err)
	}
}

func printUsage() {
	fmt.Fprint(os.Stdout, usageText)
}
