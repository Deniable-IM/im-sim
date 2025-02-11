package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
)

type Pair[A any, B any] struct {
	fst A
	snd B
}

func main() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	buildContext, err := archive.TarWithOptions(".", &archive.TarOptions{})
	if err != nil {
		panic(err)
	}

	// Server
	buildOpt := types.ImageBuildOptions{
		Dockerfile: "Dockerfile.server",
		Tags:       []string{"im-server"},
	}

	buildRes, err := cli.ImageBuild(ctx, buildContext, buildOpt)
	if err != nil {
		panic(err)
	}

	defer buildRes.Body.Close()

	io.Copy(os.Stdout, buildRes.Body)

	buildContext, err = archive.TarWithOptions(".", &archive.TarOptions{})
	if err != nil {
		panic(err)
	}

	// Client 1
	buildOpt1 := types.ImageBuildOptions{
		Dockerfile: "Dockerfile.client",
		Tags:       []string{"im-client-1"},
	}

	buildRes1, err := cli.ImageBuild(ctx, buildContext, buildOpt1)
	if err != nil {
		panic(err)
	}

	defer buildRes1.Body.Close()

	io.Copy(os.Stdout, buildRes1.Body)

	buildContext, err = archive.TarWithOptions(".", &archive.TarOptions{})
	if err != nil {
		panic(err)
	}

	// Client 2
	buildOpt2 := types.ImageBuildOptions{
		Dockerfile: "Dockerfile.client",
		Tags:       []string{"im-client-2"},
	}

	buildRes2, err := cli.ImageBuild(ctx, buildContext, buildOpt2)
	if err != nil {
		panic(err)
	}

	defer buildRes2.Body.Close()

	io.Copy(os.Stdout, buildRes2.Body)

	// Run
	networkRes, err := cli.NetworkCreate(ctx, "IMvlan", network.CreateOptions{
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
	}

	// Server
	containerRes, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "im-server",
		Tty:   true,
	}, nil, nil, nil, "im-server")
	if err != nil {
		panic(err)
	}

	// Client 1
	containerRes1, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "im-client-1",
		Tty:   true,
	}, nil, nil, nil, "im-client-1")
	if err != nil {
		panic(err)
	}

	// Client 2
	containerRes2, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "im-client-2",
		Tty:   true,
	}, nil, nil, nil, "im-client-2")
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

		if err := cli.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
			panic(err)
		}

		if err := cli.NetworkConnect(ctx, networkRes.ID, containerID, &network.EndpointSettings{
			IPAMConfig: &network.EndpointIPAMConfig{
				IPv4Address: ip,
			},
		}); err != nil {
			panic(err)
		}
	}

	fmt.Println("Container ID: ", containerRes.ID)
	fmt.Println("Network ID: ", networkRes.ID)
}
