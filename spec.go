package main

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"
)

type Specification struct {
	AppName string   `yaml:"app"`
	Version string   `yaml:"version"`
	Command *Command `yaml:"run"`
}

func newSpecification() Specification {
	return Specification{
		Command: &Command{
			isRoot: true,
		},
	}
}

func (s *Specification) StdlibPackageIsUsed(pkg string) bool {
	return s.Command.stdlibPackageIsUsed(pkg)
}

func (s *Specification) setNames() {
	s.Command.setNames(nil, s.AppName)
}

func (s *Specification) validate() error {
	if s.AppName == "" {
		return errors.New("no \"app\" defined")
	}
	if s.Command == nil {
		return errors.New("no \"run\" defined")
	}
	if err := s.Command.validate(); err != nil {
		return fmt.Errorf("%q: %w", s.AppName, err)
	}
	return nil
}

type Command struct {
	isRoot       bool
	version      string
	name         string
	parentNames  []string
	Help         string              `yaml:"help"`
	Args         []map[string]string `yaml:"args"`
	VariadicArgs *string             `yaml:"vargs"`
	Flags        map[string]*Flag    `yaml:"flags"`
	SubCommands  map[string]*Command `yaml:"subcommands"`
}

func (c *Command) IsRoot() bool {
	return c.isRoot
}

func (c *Command) Version() string {
	if c.version == "" {
		return ""
	}
	if strings.HasPrefix(c.version, "${") && strings.HasSuffix(c.version, "}") {
		inner := c.version[2 : len(c.version)-1]
		inner = strings.TrimSpace(inner)
		return os.Getenv(inner)
	}
	return c.version
}

func (c *Command) Name() string {
	return c.name
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

func (c *Command) setNames(parents []string, self string) {
	c.parentNames = parents
	c.name = self
	for childName, command := range c.SubCommands {
		command.setNames(append(parents, self), childName)
	}
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

	if len(c.SubCommands) > 0 {
		if len(c.Args) > 0 {
			return errors.New("cannot define both subcommands and args")
		}
		if c.VariadicArgs != nil {
			return errors.New("cannot define both subcommands and variadic args")
		}
	}

	for _, arg := range c.Args {
		if len(arg) > 1 {
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

	if c.VariadicArgs != nil && *c.VariadicArgs == "" {
		return errors.New("must define the name of the vargs")
	}

	for name, flag := range c.Flags {
		if err := flag.validate(); err != nil {
			return fmt.Errorf("%q: %w", name, err)
		}
	}

	return nil
}

func (c *Command) Invocation() string {
	return strings.Join(append(c.parentNames, c.name), " ")
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

func (c *Command) ArgsLen() int {
	len := 0
	for _, arg := range c.Args {
		if arg != nil {
			len++
		}
	}

	return len
}

func orderedFlagNames(flags map[string]*Flag) []string {
	var names []string
	for name := range flags {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (c *Command) Parameters(argsVarName string) string {
	var params []string
	for _, name := range orderedFlagNames(c.Flags) {
		params = append(params, fmt.Sprintf("*c.%s", toCamelCase(name)))
	}
	for i, args := range c.Args {
		if args != nil {
			params = append(params, fmt.Sprintf("%s[%d]", argsVarName, i))
		}
	}
	if c.VariadicArgs != nil {
		params = append(params, fmt.Sprintf("%s[%d:]...", argsVarName, len(c.Args)))
	}
	return strings.Join(params, ", ")
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
	if c.VariadicArgs != nil {
		argsStr = append(argsStr, fmt.Sprintf("%s ...string", *c.VariadicArgs))
	}

	return strings.Join(argsStr, ", ")
}

type Flag struct {
	Help    string        `yaml:"help"`
	Default interface{}   `yaml:"default"`
	Oneof   []interface{} `yaml:"oneof"`
}

func (f *Flag) Type() string {
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

func (f *Flag) FlagFunc() string {
	flagFuncName := strings.Split(f.Type(), ".")
	return strings.Title(flagFuncName[len(flagFuncName)-1])
}

func (f *Flag) validate() error {
	if f.Default == nil {
		return errors.New("flag default must be defined")
	}
	defaultType := reflect.TypeOf(f.Default)
	for _, choice := range f.Oneof {
		choiceType := reflect.TypeOf(choice)
		if choiceType != defaultType {
			return fmt.Errorf("%v choice in oneof must be same type as the default (%v)", choice, f.Default)
		}
	}
	return nil
}

// asArg takes an interface{} and returns it as it should be passed as a Go
// argument. For types other than durations and strings it simply returns the
// input unchanged. For strings it returns the quoted string. For durations it
// returns `time.Duration(<int64>)`.
func asArg(v interface{}) interface{} {
	if v == nil {
		// Assume it's a string.
		return "\"\""
	}
	str, ok := v.(string)
	if !ok {
		// Use the literal value.
		return v
	}
	if d, err := time.ParseDuration(str); err == nil {
		return fmt.Sprintf("time.Duration(%d)", int64(d))
	}
	return fmt.Sprintf("%q", str)
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
