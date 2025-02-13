package main

import (
	"io"
	"os"

	"github.com/docker/docker/api/types"
	dockerNetwork "github.com/docker/docker/api/types/network"

	"deniable-im/network-simulation/pkg/client"
	"deniable-im/network-simulation/pkg/container"
	"deniable-im/network-simulation/pkg/image"
	"deniable-im/network-simulation/pkg/network"
)

type Pair[A any, B any] struct {
	fst A
	snd B
}

func main() {
	client, err := client.NewClient()
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// Build server image
	serverImage := image.NewImage(*client, ".", types.ImageBuildOptions{
		Dockerfile: "Dockerfile.server",
		Tags:       []string{"im-server"},
	})

	res, err := serverImage.ImageBuild()
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	io.Copy(os.Stdout, res.Body)

	// Build client image
	clientImage1 := image.NewImage(*client, ".", types.ImageBuildOptions{
		Dockerfile: "Dockerfile.client",
		Tags:       []string{"im-client"},
	})

	res, err = clientImage1.ImageBuild()
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	io.Copy(os.Stdout, res.Body)

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
	for _, connectPair := range []Pair[*container.Container, string]{
		{server, "192.168.87.65"},
		{client1, "192.168.87.70"},
		{client2, "192.168.87.126"},
	} {
		container := connectPair.fst
		ip := connectPair.snd

		if err := container.Start(); err != nil {
			panic(err)
		}

		if err := container.NetworkConnect(network.ID, ip); err != nil {
			panic(err)
		}
	}
}
