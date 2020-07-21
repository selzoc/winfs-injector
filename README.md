# winfs-injector

## Description
This is an operator CLI tool to inject the windows file system into the Pivotal
Windows Runtime Tile.

**Contact Us**: This product is owned by the Pivotal TAS Release Engineering team. If you need support of this, you can reach us at `#tas-for-vms` on Pivotal/VMware Slack.

### Example Usage
```bash
$ winfs-injector \
  --input-tile /path/to/input.pivotal \
  --output-tile /path/to/output.pivotal
```

Note: On Windows operating systems you will need to use the bsd release of tar, which can be found [here](https://s3.amazonaws.com/bosh-windows-dependencies/tar-1503683828.exe). You should put this executable in your path as `tar.exe` before running the `winfs-injector` tool.

### Building

In order to build the winfs-injector, it needs to be built within the
[windows2016fs-release](https://github.com/cloudfoundry-incubator/windows2016fs-release).
This can be done by running the following:

```bash
git clone http://github.com/cloudfoundry-incubator/windows2016fs-release

cd windows2016fs-release

direnv allow
git submodule update --init --recursive

go get -u -v github.com/pivotal-cf/winfs-injector

cd src/github.com/pivotal-cf/winfs-injector
ginkgo -r -p

go build
```
