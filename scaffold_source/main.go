package main

import (
	"os"

	"github.com/project-name/cmd/migrate"
	"github.com/project-name/cmd/serve"
)

func main() {
	cmd := os.Args[1]

	switch cmd {
	case "migrate":
		migrate.Run()
	default:
		serve.Run()
	}
}
