package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/format"
	"os"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

//go:generate clifromyaml cli.yaml

var (
	//go:embed cli.go.tpl
	cliTemplate string
)

type application struct{}

func (application) Run(dryRun bool, outfile string, packageName string, stdout bool, yamlSpec string) error {
	b, err := os.ReadFile(yamlSpec)
	if err != nil {
		return fmt.Errorf("unable to read YAML CLI definition: %w", err)
	}

	s := newSpecification()
	if err := yaml.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("unable to unmarshal YAML definition: %w", err)
	}
	s.Command.version = s.Version
	s.setNames()

	if err := s.validate(); err != nil {
		return fmt.Errorf("error parsing cli yaml definition: %w", err)
	}

	tmpl, err := template.New("cli").Funcs(template.FuncMap{
		"toCamelCase": toCamelCase,
		"asArg":       asArg,
		"title":       strings.Title,
	}).Parse(cliTemplate)
	if err != nil {
		panic(err)
	}

	buf := new(bytes.Buffer)
	v := struct {
		*Specification
		PackageName string
	}{&s, packageName}
	if err := tmpl.Execute(buf, v); err != nil {
		return fmt.Errorf("error generating Go bindings from template: %w", err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("could not gofmt generated Go binding: %w", err)
	}

	switch {
	case dryRun:
		return nil
	case stdout:
		fmt.Println(string(formatted))
		return nil
	case outfile == "":
		outfile = fmt.Sprintf("%s.go", yamlSpec)
	}

	if err := os.WriteFile(outfile, formatted, 0666); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

func main() {
	if err := NewCLI(application{}).Run(); err != nil {
		fmt.Println(err)
	}
}
