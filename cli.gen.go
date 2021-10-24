// AUTOGENERATED -- DO NOT EDIT
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

// Application defines the entrypoints to the application logic.
type Application interface {
	Clifromyaml
}

type command struct {
	w          io.Writer
	help       string
	helpBuffer *bytes.Buffer
	flags      *flag.FlagSet
}

func newCommand(name string, help string, w io.Writer) command {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	helpBuffer := new(bytes.Buffer)
	flags.SetOutput(helpBuffer)

	return command{
		w:          w,
		help:       help,
		helpBuffer: helpBuffer,
		flags:      flags,
	}
}

// appendFlagUsage returns a function that appends a string describing the flags
// usage to the slice passed in. The usage is of the form
// `[-<flag name> <flag type>]`. For boolean flags the <flag type> is omitted.
func appendFlagUsage(usage []string) func(f *flag.Flag) {
	return func(f *flag.Flag) {
		flagArg := ""
		if typ, _ := flag.UnquoteUsage(f); typ != "" {
			flagArg = fmt.Sprintf(" <%s>", typ)
		}
		usage = append(usage, fmt.Sprintf("[-%s%s]", f.Name, flagArg))
	}
}

type Clifromyaml interface {
	Run(dryRun bool, outfile string, packageName string, stdout bool, yamlSpec string) error
}

type clifromyamlCommand struct {
	command
	version     *bool
	dryRun      *bool
	outfile     *string
	packageName *string
	stdout      *bool
	clifromyaml Clifromyaml
}

func newClifromyamlCommand(w io.Writer, clifromyaml Clifromyaml) clifromyamlCommand {
	command := newCommand("clifromyaml", "Generate Golang CLI bindings from a YAML definition.", w)
	c := clifromyamlCommand{
		command:     command,
		version:     command.flags.Bool("version", false, "print version"),
		dryRun:      command.flags.Bool("dry-run", false, "Don't write the generated Go bindings anywhere, just parse the yaml and print any errors."),
		outfile:     command.flags.String("outfile", "", "The `file` that the generated CLI bindings should be written to. If empty then they will be written to <yaml-filename>.gen.go."),
		packageName: command.flags.String("package-name", "main", "The package name to use for the generated Go bindings."),
		stdout:      command.flags.Bool("stdout", false, "Print the generated CLI bindings to stdout. Note that gofmt will not be run on the output in this case."),
		clifromyaml: clifromyaml,
	}
	c.flags.Usage = c.bufferHelp

	return c
}

func (c *clifromyamlCommand) usage() string {
	usage := []string{"Usage:", "clifromyaml"}
	c.flags.VisitAll(appendFlagUsage(usage))
	usage = append(usage, "<yaml-spec>")

	return strings.Join(usage, " ")
}

func (c *clifromyamlCommand) bufferHelp() {
	fmt.Fprintf(c.helpBuffer, "%s\n\n", c.help)
	fmt.Fprintf(c.helpBuffer, "%s\n", c.usage())

	fmt.Fprintf(c.helpBuffer, "\nArguments:\n")
	fmt.Fprintln(c.helpBuffer, "  yaml-spec: the YAML file containing the CLI definition")
	fmt.Fprintf(c.helpBuffer, "\nFlags:\n")
	c.flags.PrintDefaults()
}

func (c *clifromyamlCommand) writeHelp() error {
	_, err := c.helpBuffer.WriteTo(c.w)
	return err
}

func (c *clifromyamlCommand) writeVersion() error {
	_, err := fmt.Fprintln(c.w, "${CLIFROMYAML_VERSION}")
	return err
}

func (c *clifromyamlCommand) run(args []string) error {
	switch err := c.flags.Parse(args); err {
	case nil:
		if *c.version {
			return c.writeVersion()
		}
		args = c.flags.Args()
		if len(args) < 1 {
			return fmt.Errorf("'clifromyaml': too few arguments; expect 1, but got %d", len(args))
		}
		if len(args) > 1 {
			return fmt.Errorf("'clifromyaml': too many arguments; expect 1, but got %d", len(args))
		}
		return c.clifromyaml.Run(*c.dryRun, *c.outfile, *c.packageName, *c.stdout, args[0])
	case flag.ErrHelp:
		return c.writeHelp()
	default:
		return err
	}
}

type CLI struct {
	clifromyamlCommand clifromyamlCommand
}

func NewCLI(app Application) *CLI {
	return NewCLIWithWriter(os.Stdout, app)
}

func NewCLIWithWriter(w io.Writer, app Application) *CLI {
	return &CLI{
		clifromyamlCommand: newClifromyamlCommand(w, app),
	}
}

func (c *CLI) Run() error {
	return c.clifromyamlCommand.run(os.Args[1:])
}
