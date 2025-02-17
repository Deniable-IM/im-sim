package main

import (
	dockerTypes "github.com/docker/docker/api/types"
	dockerNetwork "github.com/docker/docker/api/types/network"

	"deniable-im/im-sim/pkg/client"
	"deniable-im/im-sim/pkg/container"
	"deniable-im/im-sim/pkg/image"
	"deniable-im/im-sim/pkg/network"
	"deniable-im/im-sim/pkg/types"
)

func main() {
	client, err := client.NewClient()
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// Build server image
	_, err = image.NewImage(*client, "./cmd/signal-sim/", dockerTypes.ImageBuildOptions{
		Dockerfile: "Dockerfile.server",
		Tags:       []string{"im-server"},
	})
	if err != nil {
		panic(err)
	}

	// Build client image
	_, err = image.NewImage(*client, "./cmd/signal-sim/", dockerTypes.ImageBuildOptions{
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
	server, err := container.NewContainer(client, "im-server", "im-server", nil)
	if err != nil {
		panic(err)
	}

	client1, err := container.NewContainer(client, "im-client", "im-client-1", nil)
	if err != nil {
		panic(err)
	}

	client2, err := container.NewContainer(client, "im-client", "im-client-2", nil)
	if err != nil {
		panic(err)
	}

	// var clientContainers []*container.Container
	// for i := range make([]int, 100) {
	// 	clientContainer, err := container.NewContainer(
	// 		client,
	// 		"im-client",
	// 		"im-client-"+string(i),
	// 		nil)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	clientContainers = append(clientContainers, clientContainer)
	// }

	// /21
	// Min x.x.0.1
	// Max x.x.7.254
	// Network reserver x.x.0.0
	// Range (broadcast) reserved  x.x.0.255
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
