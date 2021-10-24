package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
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
		return err
	}

	s := newSpecification()
	if err := yaml.Unmarshal(b, &s); err != nil {
		return err
	}
	s.Command.version = s.Version
	s.setNames()

	if err := s.validate(); err != nil {
		return fmt.Errorf("error parsing cli yaml definition: %w", err)
	}

	tmpl, err := template.New("cli").Funcs(template.FuncMap{
		"toCamelCase": toCamelCase,
		"title":       strings.Title,
	}).Parse(cliTemplate)
	if err != nil {
		return err
	}

	var w io.Writer
	switch {
	case dryRun:
		w = io.Discard
	case stdout:
		w = os.Stdout
	default:
		w = new(bytes.Buffer)
	}

	v := struct {
		*Specification
		PackageName string
	}{&s, packageName}
	if err := tmpl.Execute(w, v); err != nil {
		return fmt.Errorf("error generating Go bindings from template: %w", err)
	}

	if dryRun || stdout {
		return nil
	}

	if outfile == "" {
		filename := filepath.Base(yamlSpec)
		filename = filename[:len(filename)-len(filepath.Ext(filename))] // trim the extension
		outfile = filepath.Join(filepath.Dir(yamlSpec), fmt.Sprintf("%s.gen.go", filename))
	}

	buf := w.(*bytes.Buffer)
	if err := os.WriteFile(outfile, buf.Bytes(), 0666); err != nil {
		return err
	}

	if err := exec.Command("gofmt", "-w", outfile).Run(); err != nil {
		return fmt.Errorf("gofmt failed: %w", err)
	}

	return nil
}

func main() {
	if err := NewCLI(application{}).Run(); err != nil {
		fmt.Println(err)
	}
}
