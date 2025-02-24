package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/docker/docker/api/types"
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

	// c1 := containers[:len(containers)]

	var wg sync.WaitGroup
	for i, container := range containers {
		wg.Add(1)

		go func(c types.Container) {
			defer wg.Done()

			fmt.Printf("Stopping container - Running IMAGE: %v\n", container.Names)
			noWait := 0
			if err := cli.ContainerStop(ctx, container.ID, containertypes.StopOptions{Timeout: &noWait}); err != nil {
				panic(err)
			}
		}(container)

		if i%50 == 0 {
			wg.Wait()
		}
	}

	wg.Wait()
}
