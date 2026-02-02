package templates

import (
	"bytes"
	"fmt"
	"io/fs"
	"strings"
	"text/template"

	"github.com/arianlopezc/Trabuco/internal/config"
	"github.com/arianlopezc/Trabuco/internal/utils"
	embeddedTemplates "github.com/arianlopezc/Trabuco/templates"
)

// Engine handles template loading and execution
type Engine struct {
	fs    fs.FS
	funcs template.FuncMap
}

// NewEngine creates a new template engine with embedded templates
func NewEngine() *Engine {
	return &Engine{
		fs:    embeddedTemplates.FS,
		funcs: createFuncMap(),
	}
}

// createFuncMap returns the template functions available in all templates
func createFuncMap() template.FuncMap {
	return template.FuncMap{
		// String transformations
		"pascalCase":  utils.ToPascalCase,
		"camelCase":   utils.ToCamelCase,
		"lower":       strings.ToLower,
		"upper":       strings.ToUpper,
		"title":       utils.ToTitle,
		"replace":     strings.ReplaceAll,
		"trimSuffix":  strings.TrimSuffix,
		"trimPrefix":  strings.TrimPrefix,
		"contains":    strings.Contains,
		"hasPrefix":   strings.HasPrefix,
		"hasSuffix":   strings.HasSuffix,
		"join":        strings.Join,
		"split":       strings.Split,

		// Path transformations
		"packagePath": packageToPath,

		// Conditional helpers
		"eq":  func(a, b string) bool { return a == b },
		"ne":  func(a, b string) bool { return a != b },
		"and": func(a, b bool) bool { return a && b },
		"or":  func(a, b bool) bool { return a || b },
		"not": func(a bool) bool { return !a },

		// List helpers
		"inList": inList,
		"first":  first,
		"last":   last,
	}
}

// Execute renders a template with the given data
func (e *Engine) Execute(templatePath string, data interface{}) (string, error) {
	// Read template content
	content, err := fs.ReadFile(e.fs, templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template %s: %w", templatePath, err)
	}

	// Parse template
	tmpl, err := template.New(templatePath).Funcs(e.funcs).Parse(string(content))
	if err != nil {
		return "", fmt.Errorf("failed to parse template %s: %w", templatePath, err)
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", templatePath, err)
	}

	return buf.String(), nil
}

// ExecuteString renders a template string with the given data
func (e *Engine) ExecuteString(name, templateContent string, data interface{}) (string, error) {
	tmpl, err := template.New(name).Funcs(e.funcs).Parse(templateContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// ListTemplates returns all template files matching the given pattern
func (e *Engine) ListTemplates(dir string) ([]string, error) {
	var templates []string

	err := fs.WalkDir(e.fs, dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".tmpl") {
			templates = append(templates, path)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list templates in %s: %w", dir, err)
	}

	return templates, nil
}

// TemplateExists checks if a template file exists
func (e *Engine) TemplateExists(templatePath string) bool {
	_, err := fs.Stat(e.fs, templatePath)
	return err == nil
}

// RenderToFile is a helper that renders a template and returns the content
// along with the target file path (useful for generator)
func (e *Engine) RenderToFile(templatePath string, cfg *config.ProjectConfig) (content string, err error) {
	return e.Execute(templatePath, cfg)
}

// Helper functions for templates

// packageToPath converts a Java package to a file path
func packageToPath(pkg string) string {
	return strings.ReplaceAll(pkg, ".", "/")
}

// inList checks if an item exists in a list
func inList(list []string, item string) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}

// first returns the first element of a slice
func first(list []string) string {
	if len(list) > 0 {
		return list[0]
	}
	return ""
}

// last returns the last element of a slice
func last(list []string) string {
	if len(list) > 0 {
		return list[len(list)-1]
	}
	return ""
}
