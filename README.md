# winfs-injector

## Description
This is an operator CLI tool to inject the windows file system into the [Tanzu Application Service for Windows product](https://network.tanzu.vmware.com/products/pas-windows).

**Contact Us**: This product is owned by the Platform Provider Experience team. If you need support of this, you can reach us at `#platform-provider-experience` channel on VMware Slack.

### Example Usage
```bash
$ winfs-injector \
  --input-tile /path/to/input.pivotal \
  --output-tile /path/to/output.pivotal
```

Note: On Windows operating systems you will need to use the bsd release of tar, which can be found [here](https://s3.amazonaws.com/bosh-windows-dependencies/tar-1503683828.exe). You should put this executable in your path as `tar.exe` before running the `winfs-injector` tool.

## Building

Check our build step for detailed instructions on how to build this project.
