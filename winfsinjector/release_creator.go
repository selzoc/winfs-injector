package winfsinjector

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-cli/cmd"
	"github.com/cloudfoundry/bosh-cli/ui"
	"github.com/cloudfoundry/bosh-utils/logger"
	"github.com/cloudfoundry/hydrator/imagefetcher"
)

type ReleaseCreator struct{}

func (rc ReleaseCreator) CreateRelease(releaseName, imageName, releaseDir, tarballPath, imageTagPath, registry, version string) error {
	tagData, err := ioutil.ReadFile(imageTagPath)
	if err != nil {
		return err
	}
	imageTag := string(tagData)

	hLogger := log.New(os.Stdout, "", 0)
	releaseBlob := filepath.Join(releaseDir, "blobs", releaseName)

	// TODO: Hydrator accepts a registry argument.
	h := imagefetcher.New(hLogger, releaseBlob, imageName, imageTag, false)
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
