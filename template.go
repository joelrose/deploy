package main

import (
	_ "embed"
	"fmt"
	"html/template"
	"strings"
)

//go:embed deploy.gotmpl
var tmplFile string

var tmpl = template.Must(template.New("deploy.sh").Parse(tmplFile))

type TemplateData struct {
	RegistryName     string
	RegistryUsername string
	RegistryPassword string
	DesiredImageHash string
	ContainerName    string
	Image            string
	Host             string
	Environment      string
	ContainerPort    int
}

func RenderTemplate(t TemplateData) (string, error) {
	var commands strings.Builder
	err := tmpl.Execute(&commands, t)
	if err != nil {
		return "", fmt.Errorf("executing template: %v", err)
	}

	return commands.String(), nil
}
