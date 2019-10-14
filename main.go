package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/hydrator/hydrator"
	"github.com/cloudfoundry/bosh-cli/cmd"
	"github.com/cloudfoundry/bosh-cli/ui"
	"github.com/cloudfoundry/bosh-utils/logger"
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/winfs-injector/tile"
	"github.com/pivotal-cf/winfs-injector/winfsinjector"
)

const usageText = `winfs-injector injects the Windows 2016 root file system into the Windows 2016 Runtime Tile.

Usage: winfs-injector
  --input-tile, -i   path to input tile (example: /path/to/input.pivotal)
  --output-tile, -o  path to output tile (example: /path/to/output.pivotal)
  --registry, -r     path to docker registry (example: /path/to/registry, default: "https://registry.hub.docker.com")
  --help, -h         prints this usage information
`

func main() {
	var arguments struct {
		InputTile  string `short:"i" long:"input-tile"`
		OutputTile string `short:"o" long:"output-tile"`
		Registry   string `short:"r" long:"registry" default:"https://registry.hub.docker.com"`
		Help       bool   `short:"h" long:"help"`
	}

	_, err := jhanda.Parse(&arguments, os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	if arguments.Help {
		printUsage()
		return
	}

	var tileInjector = tile.NewTileInjector()
	var zipper = tile.NewZipper()
	var releaseCreator = ReleaseCreator{}

	wd, err := ioutil.TempDir("", "")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(wd)

	app := winfsinjector.NewApplication(releaseCreator, tileInjector, zipper)
	err = app.Run(arguments.InputTile, arguments.OutputTile, wd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprint(os.Stdout, usageText)
}

type ReleaseCreator struct{}

func (rc ReleaseCreator) CreateRelease(releaseName, imageName, releaseDir, tarballPath, imageTagPath, version string) error {
	tagData, err := ioutil.ReadFile(imageTagPath)
	if err != nil {
		return err
	}
	imageTag := string(tagData)

	h := hydrator.New(log.New(os.Stdout, "", 0), filepath.Join(releaseDir, "blobs", releaseName), imageName, imageTag, false)
	if err := h.Run(); err != nil {
		return err
	}

	releaseVersion := cmd.VersionArg{}
	if err := releaseVersion.UnmarshalFlag(version); err != nil {
		return err
	}

	l := logger.NewLogger(logger.LevelInfo)
	u := ui.NewConfUI(l)
	defer u.Flush()
	deps := cmd.NewBasicDeps(u, l)

	createReleaseOpts := &cmd.CreateReleaseOpts{
		Directory: cmd.DirOrCWDArg{Path: releaseDir},
		Version:   releaseVersion,
	}

	if tarballPath != "" {
		expanded, err := filepath.Abs(tarballPath)
		if err != nil {
			return err
		}

		createReleaseOpts.Tarball = cmd.FileArg{FS: deps.FS, ExpandedPath: expanded}
	}

	// bosh create-release adds ~7GB of temp files that should be cleaned up
	tmpDir, err := ioutil.TempDir("", "winfs-create-release")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)
	os.Setenv("HOME", tmpDir)

	createReleaseCommand := cmd.NewCmd(cmd.BoshOpts{}, createReleaseOpts, deps)
	if err := createReleaseCommand.Execute(); err != nil {
		return err
	}

	return nil
}
