# winfs-injector

## Description
This is an operator CLI tool to inject the windows file system into the Pivotal
Windows Runtime Tile.

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
