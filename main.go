package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

//go:generate clifromyaml -outfile cli.gen.go cli.yaml

var (
	//go:embed cli.go.tpl
	cliTemplate string
)

type application struct{}

func (application) Run(dryRun bool, outfile string, packageName string, yamlSpec string) error {
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

	v := struct {
		Specification
		PackageName string
	}{s, packageName}

	if dryRun {
		err := tmpl.Execute(io.Discard, v)
		if err == nil {
			fmt.Println("No problems found!")
		}
		return err
	}

	if outfile == "" {
		return tmpl.Execute(os.Stdout, v)
	}

	buf := bytes.NewBuffer(nil)
	if err := tmpl.Execute(buf, v); err != nil {
		return err
	}

	if err := os.WriteFile(outfile, buf.Bytes(), 0666); err != nil {
		return err
	}

	return exec.Command("gofmt", "-w", outfile).Run()
}

func main() {
	if err := NewCLI(application{}).Run(); err != nil {
		fmt.Println(err)
	}
}
