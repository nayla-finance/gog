package new_project

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	modulePlaceholder = "github.com/project-name"
)

func CreateProject(name string, path string, template embed.FS) error {
	fmt.Printf("🎉 Creating new project '%s' at '%s'\n", name, path)

	p := filepath.Join(path, name)

	if path == "." {
		p = "."
	}

	// create directory if path is not root
	if path != "." {
		fmt.Printf("📁 Creating project directory at '%s'\n", p)
		if err := os.MkdirAll(p, 0755); err != nil {
			return err
		}
	}

	newModule := fmt.Sprintf("github.com/%s", name)
	fmt.Printf("⚙️  Setting module name to '%s'\n", newModule)

	// Copy template files
	fmt.Println("✨ Generating files...")
	if err := copyDir(template, "template", p); err != nil {
		return err
	}

	fmt.Println("🔍 Updating import paths in files...")

	err := updateImports(p, newModule)
	if err != nil {
		return err
	}

	// In CreateProject function:
	steps := []cmdStep{
		{emoji: "🚀", name: "Initializing project", command: "go", args: []string{"mod", "init", newModule}},
		{emoji: "🔍", name: "Tidying project", command: "go", args: []string{"mod", "tidy"}},
		{emoji: "🔍", name: "Initializing git repository", command: "git", args: []string{"init"}},
	}

	if err := runCommands(p, steps); err != nil {
		return err
	}

	fmt.Println("✅ Tidied project")

	return nil
}

func copyDir(template embed.FS, src string, dst string) error {
	return fs.WalkDir(template, src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Get relative path from src
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dst, relPath)

		if d.IsDir() {
			fmt.Printf("  📂 Creating directory '%s'\n", targetPath)
			return os.MkdirAll(targetPath, 0755)
		}

		// Copy file contents
		fmt.Printf("  📄 Copying file '%s'\n", relPath)
		data, err := template.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(targetPath, data, 0644)
	})
}

func replaceInFile(path, old, new string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	newContent := strings.ReplaceAll(string(content), old, new)
	return os.WriteFile(path, []byte(newContent), 0644)
}

func updateImports(path string, newModule string) error {
	return filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error walking path '%s': %s\n", path, err)
			return err
		}

		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "go.mod") {
			return replaceInFile(path, modulePlaceholder, newModule)
		}

		return nil
	})
}

type cmdStep struct {
	emoji   string
	name    string
	command string
	args    []string
}

func runCommands(dir string, steps []cmdStep) error {
	for _, step := range steps {
		fmt.Printf("%s %s...\n", step.emoji, step.name)

		cmd := exec.Command(step.command, step.args...)
		cmd.Dir = dir // Set working directory

		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("%s failed: %w\nOutput: %s", step.name, err, output)
		}

		fmt.Printf("✅ %s complete\n", step.name)
	}
	return nil
}
