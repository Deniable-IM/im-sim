package main

import (
	"fmt"
	"sync"

	dockerTypes "github.com/docker/docker/api/types"
	dockerContainer "github.com/docker/docker/api/types/container"
	dockerNetwork "github.com/docker/docker/api/types/network"

	"deniable-im/im-sim/internal/types"
	"deniable-im/im-sim/pkg/client"
	"deniable-im/im-sim/pkg/container"
	"deniable-im/im-sim/pkg/image"
	"deniable-im/im-sim/pkg/network"
)

func main() {
	dockerClient, err := client.NewClient()
	if err != nil {
		panic(err)
	}
	defer dockerClient.Close()

	// Build server image
	_, err = image.NewImage(*dockerClient, "./cmd/signal-sim/", dockerTypes.ImageBuildOptions{
		Dockerfile: "Dockerfile.server",
		Tags:       []string{"im-server"},
	})
	if err != nil {
		panic(err)
	}

	// Build client image
	_, err = image.NewImage(*dockerClient, "./cmd/signal-sim/", dockerTypes.ImageBuildOptions{
		Dockerfile: "Dockerfile.client",
		Tags:       []string{"im-client"},
	})
	if err != nil {
		panic(err)
	}

	// Create network that supports 2046 IPs
	networkOptions := network.Options{
		Driver: "macvlan",
		IPAM: &dockerNetwork.IPAM{
			Config: []dockerNetwork.IPAMConfig{
				{
					Subnet:  "10.10.240.0/20",
					IPRange: "10.10.248.0/21",
					Gateway: "10.10.248.1",
				},
			},
		},
	}

	// Create network
	networkIMvlan := network.NewNetwork(*dockerClient, "IMvlan", networkOptions)

	// Setup Server
	serverIP := "10.10.248.2"
	server, err := container.NewContainer(
		dockerClient,
		"im-server",
		"im-server",
		&container.Options{
			Network: networkIMvlan,
			Ipv4:    &serverIP,
			HostConfig: &dockerContainer.HostConfig{
				Runtime: "crun",
			},
		})
	if err != nil {
		panic(err)
	}

	if err := server.Start(); err != nil {
		panic(err)
	}

	// Setup Clients
	var images []types.Pair[string, string]
	for i := range [100]int{} {
		images = append(images,
			types.Pair[string, string]{
				Fst: "im-client",
				Snd: fmt.Sprintf("im-client-%d", i),
			})
	}

	clientContainers, err := container.NewContainerSlice(
		dockerClient,
		images,
		&container.Options{
			HostConfig: &dockerContainer.HostConfig{
				Runtime: "crun",
			},
		})
	if err != nil {
		panic(err)
	}

	reservedIP := []string{"10.10.248.2"}
	clientContainers, err = container.AssignIP(clientContainers, reservedIP, networkOptions)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	for i, client := range clientContainers {
		wg.Add(1)

		go func(c *container.Container) {
			defer wg.Done()

			if err := c.Start(); err != nil {
				panic(err)
			}

			if err := c.NetworkConnect(networkIMvlan.ID); err != nil {
				panic(err)
			}
		}(client)

		if i%50 == 0 {
			wg.Wait()
		}
	}

	wg.Wait()
}
