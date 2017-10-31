package main

import (
	"fmt"
	"log"
	"os"

	"github.com/pivotal-cf/jhanda/flags"
	"github.com/pivotal-cf/winfs-injector/injector"
)

func main() {

	var global struct {
		InputTile string `short:"i" long:"input-tile"        description:"path to the input tile"                        default:""`
		OutputDir string `short:"o" long:"output"            description:"path to extract to"                            default:""`
	}

	_, err := flags.Parse(&global, os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	var container = new(injector.ProductionExtractContainer)
	var extractor = injector.NewExtractor(container)

	destDir, err := extractor.ExtractWindowsFSRelease(global.InputTile, global.OutputDir)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(destDir)

	// pathToScript := filepath.Join(destDir, "some-tile", "embed", "windows2016fs-release", "scripts", "create-release")
	// fmt.Println(pathToScript)
	// command := exec.Command(pathToScript)
	// err = command.Run()
	// if err != nil {
	// 	panic(err)
	// }

	// execute ./destDir/scripts/create-release
}
