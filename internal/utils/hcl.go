package utils

import (
	"bytes"
	"os"
	"text/template"
)

func HCLTemplate(data string, params map[string]any) string {
	var out bytes.Buffer
	tmpl := template.Must(template.New("hcl").Parse(data))
	err := tmpl.Execute(&out, params)
	if err != nil {
		panic(err)
	}
	return out.String()
}

func HCLTemplateFromPath(path string, params map[string]any) string {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	var out bytes.Buffer
	tmpl := template.Must(template.New("hcl").Parse(string(data)))
	err = tmpl.Execute(&out, params)
	if err != nil {
		panic(err)
	}
	return out.String()
}
