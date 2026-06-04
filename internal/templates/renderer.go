package templates

import (
	"bytes"
	"embed"
	"html/template"
)

var templateFiles embed.FS

// RenderTemplate parses any requested HTML template file and injects the provided data.
func RenderTemplate(templateName string, data interface{}) (string, error) {
	t, err := template.ParseFS(templateFiles, templateName)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
