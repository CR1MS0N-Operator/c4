// Package template provides basic text/template rendering helpers.
package template

import (
	"fmt"
	"os"
	"text/template"
)

// Render reads a template from tmplPath, executes it with vars, and writes the
// result to outputPath. Intermediate directories are created as needed.
func Render(tmplPath, outputPath string, vars map[string]string) error {
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("failed to parse template %q: %w", tmplPath, err)
	}

	if err := os.MkdirAll(osDir(outputPath), 0o755); err != nil {
		return fmt.Errorf("failed to create output directory for %q: %w", outputPath, err)
	}

	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file %q: %w", outputPath, err)
	}
	defer out.Close()

	if err := tmpl.Execute(out, vars); err != nil {
		return fmt.Errorf("failed to execute template %q: %w", tmplPath, err)
	}
	return nil
}

func osDir(path string) string {
	dir := path[:len(path)-len(fileName(path))]
	if dir == "" {
		dir = "."
	}
	return dir
}

func fileName(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == os.PathSeparator {
			return path[i+1:]
		}
	}
	return path
}
