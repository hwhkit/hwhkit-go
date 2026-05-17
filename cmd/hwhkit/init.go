package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

//go:embed templates/*
var templatesFS embed.FS

type tmplVars struct {
	ModuleName string
	ProjectName string
	GoVersion  string
}

func initCmd() *cobra.Command {
	var templateName, modulePath string
	cmd := &cobra.Command{
		Use:   "init <project-name>",
		Short: "Scaffold a new hwhkit-go project from a template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project := args[0]
			if modulePath == "" {
				modulePath = "github.com/yourorg/" + project
			}
			vars := tmplVars{ModuleName: modulePath, ProjectName: project, GoVersion: "1.23"}
			if err := scaffoldFromEmbed(templateName, project, vars); err != nil {
				return err
			}
			fmt.Printf("✓ project %q created at ./%s\n", project, project)
			fmt.Println("  next steps:")
			fmt.Println("    cd " + project + " && go mod tidy && go run .")
			return nil
		},
	}
	cmd.Flags().StringVar(&templateName, "template", "minimal-api", "template to use")
	cmd.Flags().StringVar(&modulePath, "module", "", "Go module path (default: github.com/yourorg/<name>)")
	return cmd
}

func scaffoldFromEmbed(templateName, dst string, vars tmplVars) error {
	if _, err := os.Stat(dst); err == nil {
		return fmt.Errorf("destination %q already exists", dst)
	}
	root := "templates/" + templateName
	walked := 0
	err := fs.WalkDir(templatesFS, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		walked++
		rel, _ := filepath.Rel(root, path)
		if rel == "." {
			return os.MkdirAll(dst, 0o755)
		}
		out := filepath.Join(dst, strings.TrimSuffix(rel, ".tmpl"))
		if d.IsDir() {
			return os.MkdirAll(out, 0o755)
		}
		b, err := fs.ReadFile(templatesFS, path)
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".tmpl") {
			t, err := template.New(rel).Parse(string(b))
			if err != nil {
				return err
			}
			f, err := os.Create(out)
			if err != nil {
				return err
			}
			defer f.Close()
			return t.Execute(f, vars)
		}
		return os.WriteFile(out, b, 0o644)
	})
	if err != nil {
		return err
	}
	if walked == 0 {
		return fmt.Errorf("template %q not found", templateName)
	}

	if _, err := exec.LookPath("go"); err == nil {
		_ = runIn(dst, "go", "mod", "tidy")
	}
	return nil
}

func runIn(dir string, name string, args ...string) error {
	c := exec.Command(name, args...)
	c.Dir = dir
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}
