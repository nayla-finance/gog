alias b := build

default:
	@just --list

build:
	go build -o bin/gog ./main.go
