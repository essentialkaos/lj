<p align="center"><a href="#readme"><img src=".github/images/card.svg"/></a></p>

<p align="center">
  <a href="https://kaos.sh/w/lj/ci-push"><img src="https://kaos.sh/w/lj/ci-push.svg" alt="GitHub Actions CI Status" /></a>
  <a href="https://kaos.sh/w/lj/codeql"><img src="https://kaos.sh/w/lj/codeql.svg" alt="GitHub Actions CodeQL Status" /></a>
  <a href="#license"><img src=".github/images/license.svg"/></a>
</p>

<p align="center"><a href="#installation">Installation</a> • <a href="#command-line-completion">Command-line completion</a> • <a href="#man-documentation">Man documentation</a> • <a href="#usage">Usage</a> • <a href="#ci-status">CI Status</a> • <a href="#contributing">Contributing</a> • <a href="#license">License</a></p>

<br/>

`lj` is is a tool for viewing JSON logs.

### Installation

#### From source

To build the `lj` from scratch, make sure you have a working Go 1.23+ workspace (_[instructions](https://go.dev/doc/install)_), then:

```
go install github.com/essentialkaos/lj@latest
```

#### Prebuilt binaries

You can download prebuilt binaries for Linux and macOS from [EK Apps Repository](https://apps.kaos.st/lj/latest):

```bash
bash <(curl -fsSL https://apps.kaos.st/get) lj
```

### Command-line completion

You can generate completion for `bash`, `zsh` or `fish` shell.

Bash:
```bash
sudo lj --completion=bash 1> /etc/bash_completion.d/lj
```

ZSH:
```bash
sudo lj --completion=zsh 1> /usr/share/zsh/site-functions/lj
```

Fish:
```bash
sudo lj --completion=fish 1> /usr/share/fish/vendor_completions.d/lj.fish
```

### Man documentation

You can generate man page using next command:

```bash
lj --generate-man | sudo gzip > /usr/share/man/man1/lj.1.gz
```

### Usage

<p align="center"><img src=".github/images/usage.svg"/></p>

### CI Status

| Branch | Status |
|--------|----------|
| `master` | [![CI](https://kaos.sh/w/lj/ci-push.svg?branch=master)](https://kaos.sh/w/lj/ci-push?query=branch:master) |
| `develop` | [![CI](https://kaos.sh/w/lj/ci-push.svg?branch=develop)](https://kaos.sh/w/lj/ci-push?query=branch:develop) |

### Contributing

Before contributing to this project please read our [Contributing Guidelines](https://github.com/essentialkaos/contributing-guidelines#contributing-guidelines).

### License

[Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0)

<p align="center"><a href="https://kaos.dev"><img src="https://raw.githubusercontent.com/essentialkaos/.github/refs/heads/master/images/ekgh.svg"/></a></p>
