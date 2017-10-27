package main

import (
	"log"
	"os"

	"github.com/pivotal-cf/jhanda/flags"
	"github.com/pivotal-cf/winfs-injector/injector"
)

func main() {

	var global struct {
		InputTile string `short:"i" long:"input-tile"        description:"path to the input tile"                        default:""`
	}

	_, err := flags.Parse(&global, os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	var container = new(injector.ProductionExtractContainer)
	var extractor = injector.NewExtractor(container)

	err = extractor.ExtractWindowsFSRelease(global.InputTile)
	if err != nil {
		log.Fatal(err)
	}
}
