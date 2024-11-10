package n

import (
	"embed"
	"fmt"
	"os"

	"github.com/mohamedalosaili/gog/internal/new_project"
	"github.com/spf13/cobra"
)

var NewCmd = &cobra.Command{
	Use:                   "new [project name] [path]",
	Short:                 "Create a new project",
	Args:                  cobra.ExactArgs(2),
	Aliases:               []string{"n"},
	DisableFlagsInUseLine: true,
	Run:                   runNew,
}

//go:embed template template/.*
var template embed.FS

func Run() {

	if err := NewCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runNew(cmd *cobra.Command, args []string) {
	fmt.Println("")

	name := args[0]
	path := args[1]

	fmt.Println("Creating project...")
	if err := new_project.CreateProject(name, path, template); err != nil {
		fmt.Println("Error creating project:", err)
		os.Exit(1)
	}
}
