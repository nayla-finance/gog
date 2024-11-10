default:
	@just --list


new:
	go run . new

build:
	go build -o bin/gog ./main.go
