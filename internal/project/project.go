package project

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
	modulePlaceholder = "github.com/PROJECT_NAME"
	projectName       = "PROJECT_NAME"
)

type Project struct {
	template    embed.FS
	name        string
	templateDir string
	dir         string
	repo        string
}

func NewProject(template embed.FS, name string, dir string, repo string) *Project {
	p := &Project{
		template:    template,
		name:        name,
		templateDir: "template",
		dir:         dir,
		repo:        repo,
	}

	p.dir = p.projectDir()

	return p
}

func (p *Project) Create() error {
	if info, err := os.Stat(p.dir); info != nil && err == nil {
		if err := os.MkdirAll(p.dir, 0755); err != nil {
			return fmt.Errorf("❌ Failed to create project directory '%s': %w", p.dir, err)
		}
	}

	fmt.Printf("🎉 Creating new project '%s'\n", p.name)

	newModule := ""
	if p.repo != "" {
		newModule = fmt.Sprintf("github.com/%s/%s", p.repo, p.name)
	} else {
		newModule = fmt.Sprintf("github.com/%s", p.name)
	}

	// Copy template files
	fmt.Println("✨ Creating files...")
	replaceFuncs := []func(data []byte) []byte{
		replaceInFile(modulePlaceholder, newModule), // start with module first
		replaceInFile(projectName, p.name),
	}

	if err := p.copyTemplateFiles(replaceFuncs); err != nil {
		return err
	}

	// create .env from .env.example
	envExampleFile, err := os.ReadFile(filepath.Join(p.dir, ".env.example"))
	if err != nil {
		return fmt.Errorf("❌ Failed to read .env.example file: %w", err)
	}

	envFile := filepath.Join(p.dir, ".env")
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		os.WriteFile(envFile, envExampleFile, 0644)
	}

	// In CreateProject function:
	steps := []cmdStep{
		{emoji: "🚀", name: "Initializing project", command: "go", args: []string{"mod", "init", newModule}},
		{emoji: "🔍", name: "Tidying project", command: "go", args: []string{"mod", "tidy"}},
		{emoji: "🔍", name: "Initializing git repository", command: "git", args: []string{"init"}},
	}

	if err := p.runCommands(steps); err != nil {
		return err
	}

	fmt.Println("\n\n\n✅ Project created successfully!")
	fmt.Printf("\n  To get started, run:\n\n")
	if !p.isCurrentDir() {
		fmt.Printf("  cd %s\n", p.dir)
	}
	fmt.Printf("  just serve\n\n\n")
	fmt.Println(`
    ʕ◔ϖ◔ʔ < Happy coding!
    `)
	return nil
}

func (p *Project) projectDir() string {
	// if dir is intentionally set to . it'll scaffold the project in the current directory
	// otherwise it'll create the project in the specified directory or ./project-name
	if p.dir == "" {
		return p.name
	}

	return p.dir
}

func (p *Project) copyTemplateFiles(replaceFuncs []func(data []byte) []byte) error {
	return fs.WalkDir(p.template, p.templateDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Get relative path from src
		relPath, err := filepath.Rel(p.templateDir, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(p.dir, relPath)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		// Copy file contents
		if p.isCurrentDir() {
			fmt.Printf("  📄 Creating file '%s'\n", relPath)
		} else {
			fmt.Printf("  📄 Creating file '%s/%s'\n", p.dir, relPath)
		}

		data, err := p.template.ReadFile(path)
		if err != nil {
			return err
		}

		for _, replaceFunc := range replaceFuncs {
			data = replaceFunc(data)
		}

		return os.WriteFile(targetPath, data, 0644)
	})
}

func replaceInFile(old, new string) func(content []byte) []byte {
	return func(content []byte) []byte {
		return []byte(strings.ReplaceAll(string(content), old, new))
	}
}

type cmdStep struct {
	emoji   string
	name    string
	command string
	args    []string
}

func (p *Project) runCommands(steps []cmdStep) error {
	for _, step := range steps {
		fmt.Printf("%s %s...\n", step.emoji, step.name)

		cmd := exec.Command(step.command, step.args...)
		cmd.Dir = p.dir // Set working directory

		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("%s failed: %w\nOutput: %s", step.name, err, output)
		}

		fmt.Printf("✅ %s complete\n", step.name)
	}
	return nil
}

func (p *Project) isCurrentDir() bool {
	return p.dir == "." || p.dir == "./"
}
