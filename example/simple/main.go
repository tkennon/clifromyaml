package main

import (
	"fmt"
	"time"
)

//go:generate clifromyaml cli.yaml

type simple struct {
	// Stuff
}

func (simple) Run(wait time.Duration, first string, second string, vargs ...string) error {
	fmt.Printf("Doing foo: wait=%s, first=%q, second=%q, vargs=%v\n", wait, first, second, vargs)
	return nil
}

func main() {
	if err := NewCLI(&simple{}).Run(); err != nil {
		fmt.Println(err)
	}
}
