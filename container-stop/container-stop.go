package main

import (
	"context"
	"fmt"

	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func main() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	containers, err := cli.ContainerList(ctx, containertypes.ListOptions{})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		fmt.Printf("Stopping container - Running IMAGE: %s\n", container.Image)
		noWait := 0
		if err := cli.ContainerStop(ctx, container.ID, containertypes.StopOptions{Timeout: &noWait}); err != nil {
			panic(err)
		}
	}

}
