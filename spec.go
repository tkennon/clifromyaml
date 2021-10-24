package main

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

type Specification struct {
	AppName string   `yaml:"app"`
	Version string   `yaml:"version"`
	Command *Command `yaml:"run"`
}

type Command struct {
	isRoot       bool
	version      string
	name         string
	parentNames  []string
	Help         string              `yaml:"help"`
	Args         []map[string]string `yaml:"args"`
	VariadicArgs bool                `yaml:"vargs"`
	Flags        map[string]*Flag    `yaml:"flags"`
	SubCommands  map[string]*Command `yaml:"subcommands"`
}

type Flag struct {
	Help    string      `yaml:"help"`
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

func (s Specification) setNames() {
	s.Command.setNames(nil, s.AppName)
}

func (s Specification) validate() error {
	if s.Command == nil {
		return errors.New("no \"exec:\" defined")
	}
	if err := s.Command.validate(); err != nil {
		return fmt.Errorf("%q: %w", s.AppName, err)
	}
	return nil
}

func (c *Command) IsRoot() bool {
	return c.isRoot
}

func (c *Command) Version() string {
	return c.version
}

func (c *Command) Name() string {
	return c.name
}

func (c *Command) WithName(prefix, name string) *Command {
	c.name = prefix + name
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

func (c *Command) validate() error {
	if c == nil {
		return errors.New("command cannot be null")
	}

	for name, cmd := range c.SubCommands {
		if err := cmd.validate(); err != nil {
			return fmt.Errorf("%q: %w", name, err)
		}
	}

	if len(c.SubCommands) > 0 && len(c.Args) > 0 {
		return errors.New("cannot define both args and subcommands")
	}
	if len(c.Args) > 0 && c.VariadicArgs {
		return errors.New("cannot define both args and variadic args")
	}
	if c.VariadicArgs && len(c.SubCommands) > 0 {
		return errors.New("cannot define both subcommands and variadic args")
	}

	for _, arg := range c.Args {
		if len(arg) != 1 {
			return errors.New("arg mapping must be exactly one key to one value")
		}
		for flagName := range c.Flags {
			for argName := range arg {
				if flagName == argName {
					return fmt.Errorf("cannot declare both flag and argument as %q", flagName)
				}
			}
		}
	}

	for name, flag := range c.Flags {
		if err := flag.validate(); err != nil {
			return fmt.Errorf("%q: %w", name, err)
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

func (c *Command) ParentNames() []string {
	return c.parentNames
}

func (c *Command) setNames(parents []string, self string) {
	c.parentNames = parents
	c.name = self
	for childName, command := range c.SubCommands {
		command.setNames(append(parents, self), childName)
	}
}

func (c *Command) ChainedName() string {
	if len(c.parentNames) == 0 {
		return toCamelCase(c.name)
	}

	var chainedName []string
	first := true
	for _, name := range c.parentNames {
		if first {
			first = false
			chainedName = append(chainedName, toCamelCase(name))
		} else {
			chainedName = append(chainedName, strings.Title(toCamelCase(name)))
		}
	}
	chainedName = append(chainedName, strings.Title(toCamelCase(c.name)))
	return strings.Join(chainedName, "")
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

func (f Flag) validate() error {
	if f.Default == nil {
		return errors.New("flag default must be defined")
	}
	return nil
}
