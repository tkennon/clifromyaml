package main

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

type Specification struct {
	*Command
	Version string `yaml:"version"`
}

type Command struct {
	isRoot       bool
	version      string
	Name         string
	Help         string              `yaml:"help"`
	Args         []map[string]string `yaml:"args"`
	VariadicArgs bool                `yaml:"vargs"`
	Flags        map[string]*Flag    `yaml:"flags"`
	SubCommands  map[string]*Command `yaml:"subcommands"`
}

type Flag struct {
	Help    string      `yaml:"help"`
	Dash    string      `yaml:"dash"`
	Default interface{} `yaml:"default"`
}

func toCamelCase(in string) string {
	var words []string
	for i, word := range strings.Split(in, "-") {
		if i != 0 {
			word = strings.Title(word)
		}
		words = append(words, word)
	}

	return strings.Join(words, "")
}

func newSpecification() Specification {
	return Specification{
		Command: &Command{
			isRoot: true,
		},
	}
}

func (s Specification) StdlibPackageIsUsed(pkg string) bool {
	return s.Command.stdlibPackageIsUsed(pkg)
}

func (s Specification) validate() error {
	if s.Command == nil {
		return errors.New("no \"exec:\" defined")
	}
	return s.Command.validate("exec")
}

func (c *Command) IsRoot() bool {
	return c.isRoot
}

func (c *Command) Version() string {
	return c.version
}

func (c *Command) WithName(name string) *Command {
	c.Name = name
	return c
}

func (c *Command) stdlibPackageIsUsed(pkg string) bool {
	for _, f := range c.Flags {
		if strings.HasPrefix(f.Type(), pkg+".") {
			return true
		}
	}
	for _, subCommand := range c.SubCommands {
		if subCommand.stdlibPackageIsUsed(pkg) {
			return true
		}
	}

	return false
}

func (c *Command) validate(name string) error {
	if c == nil {
		return fmt.Errorf("%q command cannot be null", name)
	}

	subCommandsDefined := false
	for name, cmd := range c.SubCommands {
		subCommandsDefined = true
		if err := cmd.validate(name); err != nil {
			return err
		}
	}

	if subCommandsDefined {
		if len(c.Args) > 0 {
			return fmt.Errorf("cannot define both args and sub-commands for %q", name)
		}
		if c.VariadicArgs {
			return fmt.Errorf("cannot define both args and variadic args for %q", name)
		}
	}

	for _, arg := range c.Args {
		if len(arg) != 1 {
			return fmt.Errorf("arg mapping for %q must be exactly one key to one value", name)
		}
		for flagName := range c.Flags {
			for argName := range arg {
				if flagName == argName {
					return fmt.Errorf("cannot declare flag and argument as %q for %q", flagName, name)
				}
			}
		}
	}

	for name, flag := range c.Flags {
		if flag.Default == nil {
			return fmt.Errorf("must define a default value for flag %q", name)
		}
	}

	return nil
}

func (c *Command) ParametersAndTypes() string {
	var argsStr []string
	for _, name := range orderedFlagNames(c.Flags) {
		argsStr = append(argsStr, fmt.Sprintf("%s %s", toCamelCase(name), c.Flags[name].Type()))
	}
	for _, arg := range c.Args {
		for argName := range arg { // We know the arg map must contain only 1 key.
			argsStr = append(argsStr, fmt.Sprintf("%s string", toCamelCase(argName)))
		}
	}
	if c.VariadicArgs {
		argsStr = append(argsStr, "vargs ...string")
	}

	return strings.Join(argsStr, ", ")
}

func orderedFlagNames(flags map[string]*Flag) []string {
	var names []string
	for name := range flags {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (c *Command) Parameters() string {
	var params []string
	for _, name := range orderedFlagNames(c.Flags) {
		params = append(params, fmt.Sprintf("*c.%s", toCamelCase(name)))
	}
	for i := range c.Args {
		params = append(params, fmt.Sprintf("args[%d]", i))
	}
	if c.VariadicArgs {
		params = append(params, "args...")
	}
	return strings.Join(params, ", ")
}

func (f Flag) Type() string {
	switch f.Default.(type) {
	case bool:
		return "bool"
	case int:
		return "int"
	case int8:
		return "int8"
	case int16:
		return "int16"
	case int32:
		return "int32"
	case int64:
		return "int64"
	case uint:
		return "uint"
	case uint8:
		return "uint8"
	case uint16:
		return "uint16"
	case uint32:
		return "uint32"
	case uint64:
		return "uint64"
	case float32:
		return "float32"
	case float64:
		return "float64"
	case string:
		if _, err := time.ParseDuration(f.Default.(string)); err == nil {
			return "time.Duration"
		}
		return "string"
	default:
		// Assume it's a string.
		return "string"
	}
}

func (f Flag) FlagFunc() string {
	flagFuncName := strings.Split(f.Type(), ".")
	return strings.Title(flagFuncName[len(flagFuncName)-1])
}

func (f Flag) DefaultArg() interface{} {
	if f.Default == nil {
		// Assume it's a string.
		return "\"\""
	}
	str, ok := f.Default.(string)
	if !ok {
		return f.Default
	}
	if d, err := time.ParseDuration(str); err == nil {
		return fmt.Sprintf("time.Duration(%d)", int64(d))
	}
	return fmt.Sprintf("%q", str)
}
