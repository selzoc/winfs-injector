package main

import (
	"log"
	"os"

	createRelease "create/createRelease"

	"github.com/pivotal-cf/jhanda/flags"
	"github.com/pivotal-cf/winfs-injector/injector"
)

func main() {
	var arguments struct {
		InputTile  string `short:"i" long:"input-tile" description:"path to the input tile" default:""`
		WorkingDir string `short:"w" long:"working-dir" description:"tmp path to work in" default:""`
	}

	_, err := flags.Parse(&arguments, os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	var container = injector.NewExtractContainer()
	var extractor = injector.NewExtractor(container)
	var releaseCreator = createRelease.ReleaseCreator{}

	app := injector.NewApplication(extractor, releaseCreator)
	err = app.Run(arguments.InputTile, arguments.WorkingDir)
	if err != nil {
		log.Fatal(err)
	}
}
