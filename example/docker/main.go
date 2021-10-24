package main

import "fmt"

//go:generate clifromyaml -outfile cli.go cli.yaml

type docker struct{}

func (*docker) RunDockerContainerLs(all bool, filter string, format string, last int, noTrunc bool, quiet bool, size bool) error {
	fmt.Printf("RunDockerContainerLs: all=%t, filter=%q, format=%q, last=%d, noTrunc=%t, quiet=%t, size=%t\n", all, filter, format, last, noTrunc, quiet, size)
	return nil
}

func (*docker) RunDockerContainerRm(force bool, link bool, volumes bool, container string, vargs ...string) error {
	fmt.Printf("RunDockerContainerRm: force=%t, link=%t, volumes=%t, container=%q, vargs=%v\n", force, link, volumes, container, vargs)
	return nil
}

func (*docker) RunDockerContainerStart(attach bool, detachKeys string, interactive bool, container string, vargs ...string) error {
	fmt.Printf("RunDockerContainerStart: attach=%t, detachKeys=%q, interactive=%t, container=%q, vargs=%v\n", attach, detachKeys, interactive, container, vargs)
	return nil
}

func (*docker) RunDockerNetworkInspect(format string, verbose bool, network string, vargs ...string) error {
	fmt.Printf("RunDockerNetworkInspect: format=%q, verbose=%t, network=%q, vargs=%v\n", format, verbose, network, vargs)
	return nil
}

func (*docker) RunDockerVolumeLs(filter string, format string, quiet bool) error {
	fmt.Printf("RunDockerVolumeLs: filter=%q, format=%q, quite=%t\n", filter, format, quiet)
	return nil
}

func (*docker) RunDockerVolumeRm(force bool, volume string, vargs ...string) error {
	fmt.Printf("RunDockerVolumeRm: force=%t, volume=%q, vargs=%v\n", force, volume, vargs)
	return nil
}

func main() {
	if err := NewCLI(&docker{}).Run(); err != nil {
		fmt.Println(err)
	}
}
