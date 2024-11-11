alias b := build

default:
	@just --list

build:
	go build -o bin/gog ./cmd
