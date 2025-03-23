package main

import (
	"fmt"

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

	// Build redis image
	_, err = image.NewImage(dockerClient, "./cmd/signal-sim/", &image.Options{
		PullOpt: &image.PullOptions{
			RefStr: "redis:latest",
		},
	})
	if err != nil {
		panic(err)
	}

	// Build DB image
	_, err = image.NewImage(dockerClient, "./cmd/signal-sim/", &image.Options{
		BuildOpt: &dockerTypes.ImageBuildOptions{
			Dockerfile: "Dockerfile.postgres",
			Tags:       []string{"im-postgres"},
		},
	})
	if err != nil {
		panic(err)
	}

	// Build server image
	_, err = image.NewImage(dockerClient, "./cmd/signal-sim/", &image.Options{
		BuildOpt: &dockerTypes.ImageBuildOptions{
			Dockerfile: "Dockerfile.server",
			Tags:       []string{"im-server"},
		},
	})
	if err != nil {
		panic(err)
	}

	// Build client image
	_, err = image.NewImage(dockerClient, "./cmd/signal-sim/", &image.Options{
		BuildOpt: &dockerTypes.ImageBuildOptions{
			Dockerfile: "Dockerfile.client",
			Tags:       []string{"im-client"},
		},
	})
	if err != nil {
		panic(err)
	}

	// Create network for DB, cache and server
	networkBackend := network.NewNetwork(dockerClient, "backend", network.Options{
		Driver: "bridge",
	})

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
	networkIMvlan := network.NewNetwork(dockerClient, "IMvlan", networkOptions)

	// Setup redis
	cache, err := container.NewContainer(
		dockerClient,
		"redis:latest",
		"im-redis",
		&container.Options{
			Connections: network.NewConnections(networkBackend),
			HostConfig: &dockerContainer.HostConfig{
				Runtime: "crun",
			},
			NetworkConfig: &dockerNetwork.NetworkingConfig{
				EndpointsConfig: map[string]*dockerNetwork.EndpointSettings{
					"backend": {
						Aliases: []string{"redis"},
					},
				},
			},
		})
	if err != nil {
		panic(err)
	}

	if err := cache.Start(); err != nil {
		panic(err)
	}

	// Setup DB
	db, err := container.NewContainer(
		dockerClient,
		"im-postgres",
		"im-postgres",
		&container.Options{
			Connections: network.NewConnections(networkBackend),
			HostConfig: &dockerContainer.HostConfig{
				Runtime: "crun",
			},
			NetworkConfig: &dockerNetwork.NetworkingConfig{
				EndpointsConfig: map[string]*dockerNetwork.EndpointSettings{
					"backend": {
						Aliases: []string{"db"},
					},
				},
			},
		})
	if err != nil {
		panic(err)
	}

	if err := db.Start(); err != nil {
		panic(err)
	}

	// Setup server
	serverIP := "10.10.248.2"
	server, err := container.NewContainer(
		dockerClient,
		"im-server",
		"im-server",
		&container.Options{
			Connections: network.NewConnections(
				networkBackend,
				network.NewAddrMapping(networkIMvlan, serverIP),
			),
			HostConfig: &dockerContainer.HostConfig{
				Runtime: "crun",
			},
			NetworkConfig: &dockerNetwork.NetworkingConfig{
				EndpointsConfig: map[string]*dockerNetwork.EndpointSettings{
					"IMvlan": {
						Aliases: []string{"server"},
					},
				},
			},
		})
	if err != nil {
		panic(err)
	}

	if err := server.Start(); err != nil {
		panic(err)
	}

	// Setup clients
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
			Connections: network.NewConnections(networkIMvlan),
			HostConfig: &dockerContainer.HostConfig{
				Runtime: "crun",
			},
		})
	if err != nil {
		panic(err)
	}

	reservedIP := []string{"10.10.248.2"}
	clientContainers, err = container.AssignIP(clientContainers, reservedIP, *networkIMvlan)
	if err != nil {
		panic(err)
	}

	container.StartContainers(clientContainers)

	// Demo
	if err := clientContainers[0].Exec([]string{"ls"}); err != nil {
		panic(err)
	}
	if err := clientContainers[1].Exec([]string{"ls", "-la"}); err != nil {
		panic(err)
	}
	if err := clientContainers[2].Exec([]string{"whoami"}); err != nil {
		panic(err)
	}
	if err := clientContainers[3].Exec([]string{"sh", "-c", "./client 1 alice"}); err != nil {
		panic(err)
	}
}
