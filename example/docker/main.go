package main

import "fmt"

//go:generate clifromyaml -outfile cli.go cli.yaml

type docker struct{}

func (*docker) RunDockerContainerLs(all bool, filter string, format string, last int, noTrunc bool, quiet bool, size bool) error {
	fmt.Printf("RunDockerContainerLs: all=%t, filter=%q, format=%q, last=%d, noTrunc=%t, quiet=%t, size=%t\n", all, filter, format, last, noTrunc, quiet, size)
	return nil
}

func (*docker) RunDockerContainerRm(force bool, link bool, volumes bool, container string, containers ...string) error {
	fmt.Printf("RunDockerContainerRm: force=%t, link=%t, volumes=%t, container=%q, containers=%v\n", force, link, volumes, container, containers)
	return nil
}

func (*docker) RunDockerContainerStart(attach bool, detachKeys string, interactive bool, container string, containers ...string) error {
	fmt.Printf("RunDockerContainerStart: attach=%t, detachKeys=%q, interactive=%t, container=%q, containers=%v\n", attach, detachKeys, interactive, container, containers)
	return nil
}

func (*docker) RunDockerNetworkInspect(format string, verbose bool, network string, networks ...string) error {
	fmt.Printf("RunDockerNetworkInspect: format=%q, verbose=%t, network=%q, networkss=%v\n", format, verbose, network, networks)
	return nil
}

func (*docker) RunDockerVolumeLs(filter string, format string, quiet bool) error {
	fmt.Printf("RunDockerVolumeLs: filter=%q, format=%q, quite=%t\n", filter, format, quiet)
	return nil
}

func (*docker) RunDockerVolumeRm(force bool, volume string, volumes ...string) error {
	fmt.Printf("RunDockerVolumeRm: force=%t, volume=%q, volumess=%v\n", force, volume, volumes)
	return nil
}

func main() {
	if err := NewCLI(&docker{}).Run(); err != nil {
		fmt.Println(err)
	}
}
