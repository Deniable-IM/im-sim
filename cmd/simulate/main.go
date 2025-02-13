package main

import (
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"

	"deniable-im/network-simulation/pkg/client"
	"deniable-im/network-simulation/pkg/image"
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
	// defer client.Cli.Close()

	// Server
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

	// Client 1
	clientImage1 := image.NewImage(*client, ".", types.ImageBuildOptions{
		Dockerfile: "Dockerfile.client",
		Tags:       []string{"im-client-1"},
	})

	res, err = clientImage1.ImageBuild()
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	io.Copy(os.Stdout, res.Body)

	// Client 2
	clientImage2 := image.NewImage(*client, ".", types.ImageBuildOptions{
		Dockerfile: "Dockerfile.client",
		Tags:       []string{"im-client-2"},
	})

	res, err = clientImage2.ImageBuild()
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	io.Copy(os.Stdout, res.Body)

	// Run
	var networkID string
	networkInspect, err := client.Cli.NetworkInspect(client.Ctx, "IMvlan", network.InspectOptions{})
	if err != nil {
		networkRes, err := client.Cli.NetworkCreate(client.Ctx, "IMvlan", network.CreateOptions{
			Driver: "macvlan",
			IPAM: &network.IPAM{
				Config: []network.IPAMConfig{
					{
						Subnet:  "192.168.87.0/24",
						IPRange: "192.168.87.64/26",
						Gateway: "192.168.87.1",
					},
				},
			},
		})
		if err != nil {
			panic(err)
		} else {
			networkID = networkRes.ID
		}
	} else {
		networkID = networkInspect.ID
	}

	// Server
	containerRes, err := client.Cli.ContainerCreate(client.Ctx, &container.Config{
		Image: "im-server",
		Tty:   true,
	}, nil, nil, nil, "")
	if err != nil {
		panic(err)
	}

	// Client 1
	containerRes1, err := client.Cli.ContainerCreate(client.Ctx, &container.Config{
		Image: "im-client-1",
		Tty:   true,
	}, nil, nil, nil, "")
	if err != nil {
		panic(err)
	}

	// Client 2
	containerRes2, err := client.Cli.ContainerCreate(client.Ctx, &container.Config{
		Image: "im-client-2",
		Tty:   true,
	}, nil, nil, nil, "")
	if err != nil {
		panic(err)
	}

	for _, con := range []Pair[string, string]{
		{containerRes.ID, "192.168.87.65"},
		{containerRes1.ID, "192.168.87.70"},
		{containerRes2.ID, "192.168.87.126"},
	} {
		containerID := con.fst
		ip := con.snd

		if err := client.Cli.ContainerStart(client.Ctx, containerID, container.StartOptions{}); err != nil {
			panic(err)
		}

		if err := client.Cli.NetworkConnect(client.Ctx, networkID, containerID, &network.EndpointSettings{
			IPAMConfig: &network.EndpointIPAMConfig{
				IPv4Address: ip,
			},
		}); err != nil {
			panic(err)
		}
	}

	fmt.Println("Container ID: ", containerRes.ID)
	fmt.Println("Network ID: ", networkID)
}
