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
go install github.com/mohamedalosaili/gog@latest
```

### Path Setup (for Linux/macOS)

After installation, ensure the Go bin directory is in your PATH:

#### For Linux/macOS
Add this line to your `~/.bashrc`, `~/.zshrc`, or equivalent shell configuration file:
```bash
export PATH=$PATH:$(go env GOPATH)/bin
```
Then reload your shell configuration:
```bash
source ~/.bashrc  # or source ~/.zshrc
```





