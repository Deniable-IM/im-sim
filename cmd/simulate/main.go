package main

import (
	dockerTypes "github.com/docker/docker/api/types"
	dockerNetwork "github.com/docker/docker/api/types/network"

	"deniable-im/network-simulation/pkg/client"
	"deniable-im/network-simulation/pkg/container"
	"deniable-im/network-simulation/pkg/image"
	"deniable-im/network-simulation/pkg/network"
	"deniable-im/network-simulation/pkg/types"
)

func main() {
	client, err := client.NewClient()
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// Build server image
	_, err = image.NewImage(*client, ".", dockerTypes.ImageBuildOptions{
		Dockerfile: "Dockerfile.server",
		Tags:       []string{"im-server"},
	})
	if err != nil {
		panic(err)
	}

	// Build client image
	_, err = image.NewImage(*client, ".", dockerTypes.ImageBuildOptions{
		Dockerfile: "Dockerfile.client",
		Tags:       []string{"im-client"},
	})
	if err != nil {
		panic(err)
	}

	// Create network
	networkOptions := network.Options{
		Driver: "macvlan",
		IPAM: &dockerNetwork.IPAM{
			Config: []dockerNetwork.IPAMConfig{
				{
					Subnet:  "192.168.87.0/24",
					IPRange: "192.168.87.64/26",
					Gateway: "192.168.87.1",
				},
			},
		},
	}
	network := network.NewNetwork(*client, "IMvlan", networkOptions)

	// Create containers
	server := container.NewContainer(*client, "im-server", "im-server", nil)
	client1 := container.NewContainer(*client, "im-client", "im-client-1", nil)
	client2 := container.NewContainer(*client, "im-client", "im-client-2", nil)

	// Run and connect containers to network
	for _, connectPair := range []types.Pair[*container.Container, string]{
		{Fst: server, Snd: "192.168.87.65"},
		{Fst: client1, Snd: "192.168.87.70"},
		{Fst: client2, Snd: "192.168.87.126"},
	} {
		container := connectPair.Fst
		ip := connectPair.Snd

		if err := container.Start(); err != nil {
			panic(err)
		}

		if err := container.NetworkConnect(network.ID, ip); err != nil {
			panic(err)
		}
	}
}
