package main

import (
	"io/ioutil"
	"log"
	"os"

	createRelease "create/createRelease"

	"github.com/pivotal-cf/jhanda/flags"
	"github.com/pivotal-cf/winfs-injector/injector"
)

func main() {
	var arguments struct {
		InputTile  string `short:"i" long:"input-tile" description:"path to the input tile" default:""`
		OutputTile string `short:"o" long:"output-tile" description:"path to write tile" default:""`
	}

	_, err := flags.Parse(&arguments, os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	var tileInjector = injector.NewTileInjector()
	var zipper = injector.NewZipper()
	var releaseCreator = createRelease.ReleaseCreator{}

	wd, err := ioutil.TempDir("", "")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(wd)

	log.Printf("working dir: %s\n", wd)

	app := injector.NewApplication(releaseCreator, tileInjector, zipper)
	err = app.Run(arguments.InputTile, arguments.OutputTile, wd)
	if err != nil {
		log.Fatal(err)
	}
}
