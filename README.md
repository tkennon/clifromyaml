# clifromyaml

Work in progress...

TODO(tk):

- write the rest of this readme
- write unit tests

`clifromyaml` is a tool to generate a Go CLI for an application. Simply define a
CLI in yaml, run the `clifromyaml` tool to generate the bindings, and then get
on with the interesting bits of coding: application logic.

## Install

`go get github.com/tkennon/clifromyaml`

## Example

Define a CLI in a yaml file
```yaml
command:
    help: This is my application to do stuff
    subcommands:
        foo:
            help: Do a foo
            args:
                - in: the input to foo
                - out: the output of foo
            flags:
                dry-run:
                    help: don't actually write to the output
                    default: false
                wait:
                    help: wait a bit before writing to the output
                    default: 5s
        bar:
            help: Do lots of bar
            vargs: true
```

Generate the stubs in `main.go` and then pass the CLI your application type
(that implements the generted `Application` interface).
```go
package main

import (
	"fmt"
	"time"
)

//go:generate clifromyaml -outfile cli.go cli.yaml

type myApplication struct {
	// Stuff
}

func (a *myApplication) RunFoo(dryRun bool, wait time.Duration, in string, out string) error {
	fmt.Printf("Doing foo: dryRun: %t, wait: %s, in: %s, out: %s\n", dryRun, wait, in, out)
	return nil
}

func (a *myApplication) RunBar(args ...string) error {
	fmt.Println("Doing bar for:", args)
	return nil
}

func main() {
	a := myApplication{}
	if err := NewCLI(&a).Run(); err != nil {
		panic(err)
	}
}
```
When built, this runs as
```shell
$ go generate
$ go build
$ ./example -h
This is my application to do stuff
Usage: ./example {bar | foo}
Sub-Commands:
  bar: Do lots of bar
  foo: Do a foo

$ ./example foo -h
Do a foo
Usage: foo [-dry-run] [-wait <duration>] <in> <out>
Arguments:
  in: the input to foo
  out: the output of foo
Flags:
  -dry-run
        don't actually write to the output
  -wait duration
        wait a bit before writing to the output (default 5s)

$ ./example foo -wait 2m first second
Doing foo: dryRun: false, wait: 2m0s, in: first, out: second

$ ./example bar a b c d e f g h i j k l m n o p
Doing bar for: [a b c d e f g h i j k l m n o p]
```

## Yaml specification

TODO