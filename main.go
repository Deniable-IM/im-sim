package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
)

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

	buildOpt := types.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Tags:       []string{"sdk-test"},
	}

	buildRes, err := cli.ImageBuild(ctx, buildContext, buildOpt)
	if err != nil {
		panic(err)
	}

	defer buildRes.Body.Close()

	io.Copy(os.Stdout, buildRes.Body)

	containerRes, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "sdk-test",
		Tty:   true,
	}, nil, nil, nil, "")
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, containerRes.ID, container.StartOptions{}); err != nil {
		panic(err)
	}

	fmt.Println(containerRes.ID)

}
