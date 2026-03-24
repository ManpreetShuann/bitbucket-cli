package output

import (
	"bytes"
	"text/template"
)

type TemplateFormatter struct {
	Template string
}

func (f *TemplateFormatter) Format(data any) (string, error) {
	tmpl, err := template.New("output").Parse(f.Template)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
