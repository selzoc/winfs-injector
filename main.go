package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	createRelease "github.com/cloudfoundry-incubator/windows2016fs-release/src/create/createRelease"
	"github.com/pivotal-cf/jhanda/flags"
	"github.com/pivotal-cf/winfs-injector/injector"
)

func main() {
	var arguments struct {
		InputTile  string `short:"i" long:"input-tile"       description:"path to the input tile"                        default:""`
		WorkingDir string `short:"w" long:"working-dir"          description:"tmp path to work in"                           default:""`
		// Output     string `short:"o" long:"output"           description:"path to extract to"                            default:"."`
	}

	_, err := flags.Parse(&arguments, os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	var container = new(injector.ProductionExtractContainer)
	var extractor = injector.NewExtractor(container)

	destDir, err := extractor.ExtractWindowsFSRelease(arguments.InputTile, arguments.WorkingDir)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(destDir)

	imageName := "cloudfoundry/windows2016fs"

	tileName := path.Base(arguments.InputTile)
	tileName = strings.Replace(tileName, path.Ext(tileName), "", 1)

	releaseDir := filepath.Join(destDir, tileName, "embed", "windows2016fs-release")

	imageTagPath := filepath.Join(releaseDir, "src", "code.cloudfoundry.org", "windows2016fs", "IMAGE_TAG")

	versionDataPath := filepath.Join(releaseDir, "VERSION")
	var outputDir string
	outputDir = filepath.Join(releaseDir, "blobs", "windows2016fs")
	tarballPath := filepath.Join(outputDir, "windows2016fs-0.0.26.tgz")
	createRelease.CreateRelease(imageName, releaseDir, tarballPath, imageTagPath, versionDataPath, outputDir)

	// pathToScript := filepath.Join(destDir, "some-tile", "embed", "windows2016fs-release", "scripts", "create-release")
	// fmt.Println(pathToScript)
	// command := exec.Command(pathToScript)
	// err = command.Run()
	// if err != nil {
	// 	panic(err)
	// }

	// execute ./destDir/scripts/create-release
}
