{{define "command.tmpl"}}
{{if eq (len .SubCommands) 0}}type {{title (.ChainedName)}} interface {
    Run{{if not .IsRoot}}{{title .ChainedName}}{{end}}({{.ParametersAndTypes}}) error
}
{{end}}
type {{.ChainedName}}Command struct {
    command{{if ne .Version ""}}
    version *bool{{end}}
    {{range $fname, $flag := .Flags}}{{toCamelCase $fname}} *{{$flag.Type}}
    {{toCamelCase $fname}}Choices []{{$flag.Type}}
    {{end}}{{range $cname, $command := .SubCommands}}{{toCamelCase $cname}} {{$command.ChainedName}}Command
    {{else}}{{toCamelCase .Name}} {{title .ChainedName}}{{end}}
}

func new{{title .ChainedName}}Command(w io.Writer, {{if eq (len .SubCommands) 0}}{{toCamelCase .Name}} {{title .ChainedName}}{{else}}app Application{{end}}) {{.ChainedName}}Command {
    command := newCommand("{{.Name}}", "{{.Help}}", w)
    c := {{.ChainedName}}Command{
        command: command,{{if ne .Version ""}}
        version: command.flags.Bool("version", false, "print version"),{{end}}
        {{range $fname, $flag := .Flags}}{{toCamelCase $fname}}: command.flags.{{$flag.FlagFunc}}("{{$fname}}", {{asArg $flag.Default}}, "{{$flag.Help}}"),
        {{toCamelCase $fname}}Choices: {{if gt (len $flag.Oneof) 0}}[]{{$flag.Type}}{ {{range $flag.Oneof}}{{asArg .}},{{end}} }{{else}}nil{{end}},
        {{end}}{{range $cname, $command := .SubCommands}}{{toCamelCase $cname}}: new{{title $command.ChainedName}}Command(w, app),
        {{else}}{{toCamelCase .Name}}: {{toCamelCase .Name}},{{end}}
    }
    c.flags.Usage = c.bufferHelp

    return c
}

func (c *{{.ChainedName}}Command) validateFlags() error {
    {{range $fname, $flag := .Flags}}if err := func() error {
        if len(c.{{toCamelCase $fname}}Choices) == 0 {
            return nil
        }
        for _, choice := range c.{{toCamelCase $fname}}Choices {
            if choice == *c.{{toCamelCase $fname}} {
                return nil
            }
        }
        return fmt.Errorf("'{{$fname}}' must be one of %v", c.{{toCamelCase $fname}}Choices)
    }(); err != nil {
        return err
    }
    {{end}}return nil
}

func (c *{{.ChainedName}}Command) validateArgs(args []string) error {
    if len(args) < {{.ArgsLen}} {
        return fmt.Errorf("'{{.Invocation}}': too few arguments; expect {{.ArgsLen}}{{if .VariadicArgs}} or more{{end}}, but got %d", len(args))
    }
    {{if not .VariadicArgs}}if len(args) > {{.ArgsLen}} {
        return fmt.Errorf("'{{.Invocation}}': too many arguments; expect {{.ArgsLen}}, but got %d", len(args))
    }{{end}}
    return nil
}

func (c *{{.ChainedName}}Command) usage() string {
    usage := []string{"Usage:", "{{.Invocation}}"}{{if gt (len .SubCommands) 0}}
    usage = append(usage, "<command>"){{end}}
    c.flags.VisitAll(appendFlagUsage(usage))
    {{range .Args}}{{range $aname, $arg := .}}usage = append(usage, "<{{$aname}}>"){{end}}
    {{end}}{{with .VariadicArgs}}usage = append(usage, "[<{{.}}>...]"){{end}}
    return strings.Join(usage, " ")
}

func (c *{{.ChainedName}}Command) bufferHelp() {
    fmt.Fprintf(c.helpBuffer, "%s\n\n", c.help)
    fmt.Fprintf(c.helpBuffer, "%s\n", c.usage())
    {{if gt (len .SubCommands) 0}}fmt.Fprintln(c.helpBuffer, "\nCommands:"){{end}}
    {{range $cname, $command := .SubCommands}}fmt.Fprintln(c.helpBuffer, "  {{$cname}}: {{$command.Help}}")
    {{end}}
    {{if gt .ArgsLen 0}}fmt.Fprintf(c.helpBuffer, "\nArguments:\n"){{end}}
    {{range .Args}}{{range $argName, $help := .}}fmt.Fprintln(c.helpBuffer, "  {{$argName}}: {{$help}}"){{end}}
    {{end}}{{if gt (len .Flags) 0}}fmt.Fprintf(c.helpBuffer, "\nFlags:\n")
	c.flags.PrintDefaults(){{end}}
}

func (c *{{.ChainedName}}Command) writeHelp() error {
    _, err := c.helpBuffer.WriteTo(c.w)
    return err
}

{{if ne .Version ""}}
func (c *{{.ChainedName}}Command) writeVersion() error {
    _, err := fmt.Fprintln(c.w, "{{.Version}}")
    return err
}
{{end}}

func (c *{{.ChainedName}}Command) run(args []string) error {
    {{if gt (len .SubCommands) 0}}if len(args) == 0 {
        return fmt.Errorf("sub-command required")
    }
    switch args[0] { {{range $cname, $command := .SubCommands}}
    case "{{$cname}}":
        return c.{{toCamelCase $cname}}.run(args[1:]){{end}}
    default:
        err := c.flags.Parse(args)
        if err == flag.ErrHelp {
            return c.writeHelp()
        }
        if err != nil {
            return err
        }{{if ne .Version ""}}
        if *c.version {
            return c.writeVersion()
        }{{end}}
        return fmt.Errorf("Unknown command: %q", args[0])
    }{{else}}switch err := c.flags.Parse(args); err {
    case nil:
        {{if ne .Version ""}}if *c.version {
            return c.writeVersion()
        }{{end}}
        // Check that all flags are oneof the defined choices.
        if err := c.validateFlags(); err != nil {
            return err
        }
        args = c.flags.Args()
        if err := c.validateArgs(args); err != nil {
            return err
        }
        return c.{{toCamelCase .Name}}.Run{{if not .IsRoot}}{{title .ChainedName}}{{end}}({{.Parameters "args"}})
    case flag.ErrHelp:
        return c.writeHelp()
    default:
        return err
    }{{end}}
}

{{range $cname, $command := .SubCommands}}
{{template "command.tmpl" $command}}
{{end}}
{{end}}
{{define "application.tmpl"}}{{range $cname, $command := .SubCommands}}{{template "application.tmpl" $command}}{{else}}{{title .ChainedName}}
{{end}}{{end}}
// AUTOGENERATED -- DO NOT EDIT
package {{.PackageName}}

import (
    "bytes"
    "flag"
    "fmt"
    "io"
    "os"
    "strings"
    {{if .StdlibPackageIsUsed "time"}}"time"{{end}}
)

// Application defines the entrypoints to the application logic.
type Application interface {
    {{template "application.tmpl" .Command}}
}

type command struct {
    w io.Writer
    help string
    helpBuffer *bytes.Buffer
    flags *flag.FlagSet
}

func newCommand(name string, help string, w io.Writer) command {
    flags := flag.NewFlagSet(name, flag.ContinueOnError)
    helpBuffer := new(bytes.Buffer)
    flags.SetOutput(helpBuffer)

    return command{
        w: w,
        help: help,
        helpBuffer: helpBuffer,
        flags: flags,
    }
}

// appendFlagUsage returns a function that appends a string describing the flags
// usage to the slice passed in. The usage is of the form
// `[-<flag name> <flag type>]`. For boolean flags the <flag type> is omitted.
func appendFlagUsage(usage []string) func(f *flag.Flag) {
    return func( f *flag.Flag) {
        flagArg := ""
        if typ, _ := flag.UnquoteUsage(f); typ != "" {
            flagArg = fmt.Sprintf(" <%s>", typ)
        }
        usage = append(usage, fmt.Sprintf("[-%s%s]", f.Name, flagArg))
    }
}

{{template "command.tmpl" .Command}}

type CLI struct {
    {{toCamelCase .AppName}}Command {{toCamelCase .AppName}}Command
}

func NewCLI(app Application) *CLI {
    return NewCLIWithWriter(os.Stdout, app)
}

func NewCLIWithWriter(w io.Writer, app Application) *CLI {
    return &CLI{
        {{toCamelCase .AppName}}Command: new{{title (toCamelCase .AppName)}}Command(w, app),
    }
}

func (c *CLI) Run() error {
    return c.{{toCamelCase .AppName}}Command.run(os.Args[1:])
}
