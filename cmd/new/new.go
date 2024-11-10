package n

import (
	"embed"
	"fmt"
	"os"

	"github.com/mohamedalosaili/gog/internal/project"
	"github.com/spf13/cobra"
)

//go:embed template template/.*
var template embed.FS

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new [project name]",
		Short: "Create a new project",
		Args:  cobra.ExactArgs(1),
		RunE:  runNew,
	}

	cmd.Flags().StringP("directory", "d", "", "The path to create the project in (e.g. ./my-project)")
	cmd.Flags().StringP("username", "u", "", "Github username to create the project in (e.g. github.com/your-github-username/project-name)")

	return cmd
}

func runNew(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("❌ Missing project name")
	}

	name := args[0]
	path, err := cmd.Flags().GetString("directory")
	if err != nil {
		return fmt.Errorf("❌ Failed to get directory flag: %w", err)
	}

	gitHubUsername, err := cmd.Flags().GetString("username")
	if err != nil {
		return fmt.Errorf("❌ Failed to get username flag: %w", err)
	}

	fmt.Println(`
    ______      ______       ______   
   /      \    /      \     /      \  
  /$$$$$$  |  /$$$$$$  |   /$$$$$$  | 
  $$ | _$$/   $$ |  $$ |   $$ | _$$/ 
  $$ |/    |  $$ |  $$ |   $$ |/    | 
  $$ |$$$$ |  $$ |  $$ |   $$ |$$$$ | 
  $$ \__$$ |  $$ \__$$ |   $$ \__$$ | 
  $$    $$/   $$    $$/    $$    $$/ 
   $$$$$$/     $$$$$$/      $$$$$$/  
  `)

	p := project.NewProject(template, name, path, gitHubUsername)

	if err := p.Create(); err != nil {
		fmt.Println("Error creating project:", err)
		os.Exit(1)
	}

	return nil
}
