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
