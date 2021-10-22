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

var (
	//go:embed cli.tmpl
	cliTemplate string
)

// func main() {
// 	var s Specification
// 	b, err := os.ReadFile(os.Args[1])
// 	if err != nil {
// 		panic(err)
// 	}
// 	if err := yaml.Unmarshal(b, &s); err != nil {
// 		panic(err)
// 	}

// 	if err := s.validate(); err != nil {
// 		panic(err)
// 	}

// 	tmpl, err := template.New("cli").Funcs(template.FuncMap{
// 		"toCamelCase": toCamelCase,
// 		"title":       strings.Title,
// 	}).Parse(cliTemplate)
// 	if err != nil {
// 		panic(err)
// 	}
// 	buf := new(bytes.Buffer)
// 	if err := tmpl.Execute(buf, struct {
// 		Specification
// 		PackageName string
// 	}{s, "main"}); err != nil {
// 		panic(err)
// 	}
// 	os.Stdout.Write(buf.Bytes())
// }

func main() {
	if err := NewCLI(application{}).Run(); err != nil {
		fmt.Println(err)
	}
}

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

	if err := s.validate(); err != nil {
		return err
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
		return tmpl.Execute(io.Discard, v)
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

	format := exec.Command("gofmt", "-w", outfile)
	return format.Run()
}
