package main

import "fmt"

//go:generate clifromyaml -outfile cli.go cli.yaml

type docker struct{}

func (*docker) RunDockerContainerLs(all bool, filter string, format string, last int, noTrunc bool, quiet bool, size bool) error {
	fmt.Printf("RunDockerContainerLs: all=%t, filter=%q, format=%q, last=%d, noTrunc=%t, quiet=%t, size=%t\n", all, filter, format, last, noTrunc, quiet, size)
	return nil
}

func (*docker) RunDockerContainerRm(force bool, link bool, volumes bool, vargs ...string) error {
	fmt.Printf("RunDockerContainerRm: force=%t, link=%t, volumes=%t, vargs=%v\n", force, link, volumes, vargs)
	return nil
}

func (*docker) RunDockerContainerStart(attach bool, detachKeys string, interactive bool, vargs ...string) error {
	fmt.Printf("RunDockerContainerStart: attach=%t, detachKeys=%q, interactive=%t, vargs=%v\n", attach, detachKeys, interactive, vargs)
	return nil
}

func (*docker) RunDockerNetworkInspect(format string, verbose bool) error {
	fmt.Printf("RunDockerNetworkInspect: format=%q, verbose=%t\n", format, verbose)
	return nil
}

func (*docker) RunDockerVolumeLs(filter string, format string, quiet bool) error {
	fmt.Printf("RunDockerVolumeLs: filter=%q, format=%q, quite=%t\n", filter, format, quiet)
	return nil
}

func (*docker) RunDockerVolumeRm(force bool, vargs ...string) error {
	fmt.Printf("RunDockerVolumeRm: force=%t, vargs=%v\n", force, vargs)
	return nil
}

func main() {
	if err := NewCLI(&docker{}).Run(); err != nil {
		fmt.Println(err)
	}
}
