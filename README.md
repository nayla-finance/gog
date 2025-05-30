# gog - Go Generator CLI

A command-line tool to quickly scaffold new Go projects with a standardized structure and configuration.

## Features

- Creates a new Go project with a single command
- Sets up proper module naming and directory structure 
- Initializes git repository
- Configures go modules automatically
- Installs dependencies

## Installation

You can install gog directly using Go:

```bash
go install github.com/nayla-finance/gog/cmd/gog@latest
```

### Path Setup (for Linux/macOS)

After installation, ensure the Go bin directory is in your PATH:

Add this line to your `~/.bashrc`, `~/.zshrc`, or equivalent shell configuration file:
open `~/.bashrc`/`~/.zshrc` file
```bash
nano ~/.bashrc
#or
nano ~/.zshrc
```

Add this line to the file:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

## Usage

```bash
gog new <project-name>
```


### Troubleshooting

- If you encounter any issues, while updating the version try this:

```bash
go clean -modcache
rm -rf $GOPATH/pkg/mod/cache/download/github.com/nayla-finance/gog
go install github.com/nayla-finance/gog/cmd/gog@v<version>
```


