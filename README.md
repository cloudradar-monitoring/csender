## How to build from sources
- [Install Golang 1.9 or newer](https://golang.org/dl/)
```bash
go get -d -u github.com/cloudradar-monitoring/csender
go build -o -ldflags="-X main.version=$(git --git-dir=src/github.com/cloudradar-monitoring/csender/.git describe --always --long --dirty --tag)" csender github.com/cloudradar-monitoring/csender/cmd/csender
```

## How to run

```bash
./csender -c csender.conf -k my.key -s 1 -o foo=bla -o bytes.received=23008 -o bytes.sent=98776654 -m "All went well"
```

## Configuration
Check the [example config](https://github.com/cloudradar-monitoring/csender/blob/master/example.config.toml)

Default locations:
* Mac OS: `~/.csender/csender.conf`
* Windows: `./csender.conf`
* UNIX: `/etc/csender/csender.conf`

– Hub URL can be also set via `CSENDER_HUB_URL` environment variable (`hub_url` in config will have precendence)

– Hub Proxy can be also set via `HTTP_PROXY` environment variable (`hub_proxy` in config will have precendence)
## Logs location
* Mac OS: `~/.csender/csender.log`
* Windows: `./csender.log`
* UNIX: `/etc/csender/csender.conf`

## Build binaries and deb/rpm packages
– Install [goreleaser](https://goreleaser.com/introduction/)
```bash
goreleaser --snapshot
```