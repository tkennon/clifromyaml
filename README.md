# clifromyaml

`clifromyaml` is a tool to generate a Go CLI for an application. Simply define a
CLI in yaml, run the `clifromyaml` tool to generate the bindings, and then get
on with the interesting bits of coding: application logic.

## Install

`go get github.com/tkennon/clifromyaml` and make sure that `~/go/bin` is in your
PATH.

## Example

Define a CLI in a yaml file

```yaml
# Declares the name of the application as it will be invoked by a user
# (required).
app: example
# Declares the top level command (required). Commands consist of arguments,
# variadic arguments, flags, and sub-commands. A command may declare either a of
# subcommands, or a combination of args, vargs, and flags.
run:
  # The help string for the command. Will be printed whenever the user asks for
  # help through the automatically generated -h or --help flags.
  help: This is my application to do stuff
  # This example application declares two sub-commands: foo and bar.
  subcommands:
    foo:
      help: Do a foo
      # Declares that foo takes exactly two ordered arguments. The generated Go
      # code will refer to these arguments as `in` and `out` respectively. Both
      # args have an associated description which will appear in the printed
      # help output and example usage.
      args:
        - in: the input to foo
        - out: the output of foo
      # Declares the foo can optionally take two flags. Note that the generated
      # Go code uses the stdlib "flag" package and so the flag names may be
      # prefixed with either single or double dashes by the user (for example
      # `--wait 2m3s` or `-wait=1s`). As a consequence, single letter aliases
      # are not supported (`-w 1s`).
      flags:
        dry-run:
          # As with commands and arguments, flags have help strings. The better
          # the help strings, the easier the application will be to use.
          help: don't actually write to the output
          # A default must be decalred: this is how clifromyaml infers the type
          # of the flag. Integer, boolean, string, and time.Duration types are
          # supported.
          default: false
        wait:
          help: wait a bit before writing to the output
          default: 5s
    bar:
      # The bar command takes at least one argument, but may take a variadic
      # number.
      help: Do lots of bar
      args:
       - first: the first bar
      # The description of the vargs appears in example usage generated in the
      # printed help output.
      vargs: bars
      flags:
        # Optionally takes a flag baz that must be one of the predefined
        # choices. If it is not then an error is returned.
        baz:
          help: some optional extra baz
          default: red
          oneof: [red, blue, yellow]
```

Generate the CLI by running `clifromyaml path/to/cli.yaml`. This will create a
file called `path/to/cli.yaml.go` that contains (among other things) an
`Application` interface.

```golang
// Application defines the entrypoints to the application logic.
type Application interface {
	ExampleBar
	ExampleFoo
}

type ExampleBar interface {
	RunExampleBar(baz string, first string, bars ...string) error
}

type ExampleFoo interface {
	RunExampleFoo(dryRun bool, wait time.Duration, in string, out string) error
}
```

Then, simply implement a type that satisfies the `Application` interface, and
run the CLI with it.

```go
package main

import (
	"fmt"
	"time"
)

//go:generate clifromyaml cli.yaml

type myApplication struct {
	// Stuff
}

func (a *myApplication) RunExampleFoo(dryRun bool, wait time.Duration, in string, out string) error {
	fmt.Printf("Doing foo: dryRun: %t, wait: %s, in: %s, out: %s\n", dryRun, wait, in, out)
	return nil
}

func (a *myApplication) RunExampleBar(baz string, first string, bars ...string) error {
	fmt.Printf("Doing bar for %s and %v\n", first, bars)
	return nil
}

func main() {
	a := myApplication{}
	if err := NewCLI(&a).Run(); err != nil {
		fmt.Println(err)
	}
}
```

When built this runs as

```shell
$ ./example -h
This is my application to do stuff

Usage: example <command>

Commands:
  bar: Do lots of bar
  foo: Do a foo
```

```shell
$ ./example foo -h
Do a foo

Usage: example foo [--dry-run] [--wait <duration>] <in> <out>

Arguments:
  in: the input to foo
  out: the output of foo

Flags:
  -dry-run
        don't actually write to the output
  -wait duration
        wait a bit before writing to the output (default 5s)
```

```shell
$ ./example foo --wait 2m first second
Doing foo: dryRun: false, wait: 2m0s, in: first, out: second
```

```shell
$ ./example bar --baz pink first
'baz' must be one of [red blue yellow]
```

```shell
$ ./example bar a b c d e f g h i j k l m n o p
Doing bar for a and [b c d e f g h i j k l m n o p]
```

Check out the  `example/` directory for further examples. Also checkout
`cli.yaml` which defines the CLI used for `clifromyaml` itself.

## Usage

```shell
$ clifromyaml --help
Generate Golang CLI bindings from a YAML definition.

Usage: clifromyaml [--dry-run] [--outfile <file>] [--package-name <string>] [--stdout] [--version] <yaml-spec>

Arguments:
  yaml-spec: the YAML file containing the CLI definition

Flags:
  -dry-run
        Don't write the generated Go bindings anywhere, just parse the yaml and print any errors.
  -outfile file
        The file that the generated CLI bindings should be written to. If empty then they will be written to <yaml-spec>.go.
  -package-name string
        The package name to use for the generated Go bindings. (default "main")
  -stdout
        Print the generated CLI bindings to stdout.
  -version
        print version
```

## Yaml specification

```yaml
app: <app>
version: <version>
run:
  help: <help>
  subcommands:
  args:
   - <name>: <help>
  vargs: <name>
  flags:
    <flag name>:
      help: <help>
      default: <default>
      oneof: [<choices, ...>]
```

### app

`app` is required: it is the name of the application as it will be used in a
shell, e.g. `go`, `git`, `clifromyaml` etc.

### version

A special `--version` flag will be automatically added if the top-level
`version` key is specified in the yaml configuration, and `myapp --version` will
print the value of the `version` key to stdout.

e.g.

```yaml
app: foo
version: 1.2.3
run:
  help: Simply prints a version
```

will generate an application that does:

```shell
$ ./foo --version
1.2.3
```

You can specify environment variables using the `${}` syntax:

```yaml
app: foo
version: ${FOO_VERSION}
run:
  help: Print the version in the environment at build time
```

```shell
$ FOO_VERSION=3.2.1 go generate; go build
$ ./foo --version
3.2.1
```

### run

`run` is the entrypoint to the application; it defines what happens when the
application is run.

#### help

`help` is the string that will be printed if the application is invoked with
`-h` or `--help`. A well documented application is an easy to use application.

#### args

`args` declares a list of key:value pairs describing positional arguments that
the command expects. The keys are the names of the arguments, and the values are
a description, e.g. `- config: The configuration file`. If the application is
invoked with more or less than the exact number of arguments specified then a
descriptive error is returned.

#### vargs

`vargs` declares that the command takes a variadic number of arguments after the
poisitonal arguments (if `args` is empty then everything after the command is
parsed as a varg).

#### flags

`flags` declares a set of flags the command optionally accepts:

- `help` is a help string for the flag
- `default` defines the default value of the flag. This must be specified so
  that `clifromyaml` can infer the type of the flag.
- `oneof` optionally specifies that the flag must be set to one of the given
  choices. If invoked with any other value then an error will be returned.

Flags must be passed before any arguments: `./myapp [flags] <args>`.

#### subcommands

`subcommands` allows you to recursively define subcommands (i,e, `test` and
`build` are subcommands of `go`). You may not specify args, vargs, or flags as
well as `subcommands`.

## TODO

- write unit tests
