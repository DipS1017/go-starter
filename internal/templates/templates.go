package templates

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
)

//go:embed *.html
var templateFiles embed.FS

var parsedTemplates *template.Template

func init() {
	parsedTemplates = template.Must(parseTemplates())
}

func parseTemplates() (*template.Template, error) {
	// Add Funcs here with safeHTML function
	tmpl := template.New("").Funcs(template.FuncMap{
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
	})

	err := fs.WalkDir(templateFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		fileBytes, err := fs.ReadFile(templateFiles, path)
		if err != nil {
			return fmt.Errorf("reading template file %s: %w", path, err)
		}

		_, err = tmpl.New(path).Parse(string(fileBytes))
		if err != nil {
			return fmt.Errorf("parsing template %s: %w", path, err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking template directory: %w", err)
	}

	return tmpl, nil
}

func LoadAndExecuteTemplate(name string, data interface{}) (string, error) {
	var buf bytes.Buffer

	// Execute the specific template by its name (which should match the filename)
	err := parsedTemplates.ExecuteTemplate(&buf, name, data)
	if err != nil {
		return "", fmt.Errorf("executing template %q: %w", name, err)
	}

	return buf.String(), nil
}

func GetTemplate(name string) *template.Template {
	return parsedTemplates.Lookup(name)
}
