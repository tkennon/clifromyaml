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
