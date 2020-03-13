package winfsinjector

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/hydrator/imagefetcher"
	"github.com/cloudfoundry/bosh-cli/cmd"
	"github.com/cloudfoundry/bosh-cli/ui"
	"github.com/cloudfoundry/bosh-utils/logger"
)

type ReleaseCreator struct{}

func (rc ReleaseCreator) CreateRelease(releaseName, imageName, releaseDir, tarballPath, imageTag, registry, version string) error {
	hLogger := log.New(os.Stdout, "", 0)
	releaseBlob := filepath.Join(releaseDir, "blobs", releaseName)

	h := imagefetcher.New(hLogger, releaseBlob, imageName, imageTag, registry, false)
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
