## Why was this tool created?

It all started one day when I was looking for how homebrew saves their bottles, according to the official documentation I found out that they save them in github packages, then I continued my research and found out that they package their software in tar.gz and upload it as OCI artifacts to github packages.

Then I wondered if I could use it as a package storage on [Azathothas/Toolpacks](https://github.com/Azathothas/Toolpacks/)

So [@Azathothas](https://github.com/Azathothas) asked if I could write a tool to do this.

That's how this tool was born.

## Usage

```shell
go install github.com/akkuman/blob-uploader/cmd/...
```

```
$ ./blob-uploader upload -h
upload file bolb to github packages

ref: https://github.com/Homebrew/brew/blob/b753315b0b1e78b361612bf4985502bf9dca5582/Library/Homebrew/github_packages.rb#L196-L428

Usage:
  blob-uploader upload [flags]

Flags:
  -h, --help                  help for upload
      --image-source string   value of org.opencontainers.image.source, if blank, default to current repo url
  -p, --password string       the password of registry
      --platform string       Specify platform (e.g. linux/amd64) (default "linux/amd64")
  -r, --ref-name string       the ref that you will push (e.g. ghcr.io/example/hello:1.2.0)
  -f, --tgz-file string       file path for tgz which will be uploaded
  -u, --username string       the username of registry
```

it also supports config from environment, for example, the above command line arguments can be replaced with the following environment variables.

```shell
export BUL_USERNAME=akkuman
export BUL_PASSWORD=<ghp_token>
export BUL_REF_NAME=ghcr.io/example/hello:1.2.0
export BUL_TGZ_FILE=/tmp/example.tgz
```
