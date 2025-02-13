package container

import (
	"deniable-im/network-simulation/pkg/client"
	"errors"
	"fmt"
	"log"

	dockerContainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	dockerNetwork "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/errdefs"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

type Options struct {
	containerConfig *dockerContainer.Config
	hostConfig      *dockerContainer.HostConfig
	networkConfig   *dockerNetwork.NetworkingConfig
	platform        *v1.Platform
}

type Container struct {
	Client client.Client
	ID     string
	Image  string
	Name   string
}

func NewContainer(client client.Client, image string, name string, options *Options) *Container {
	if options == nil {
		options = &Options{
			&dockerContainer.Config{},
			&dockerContainer.HostConfig{},
			&dockerNetwork.NetworkingConfig{},
			&v1.Platform{},
		}
	}
	options.containerConfig.Image = image

	res, err := client.Cli.ContainerCreate(
		client.Ctx,
		options.containerConfig,
		options.hostConfig,
		options.networkConfig,
		options.platform,
		name)
	if err != nil {
		if errors.Is(err, errdefs.Conflict(err)) {
			log.Printf("Container %s is already in use.", name)
			id, err := GetIdByName(client, name)
			if err != nil {
				panic(err)
			}
			return &Container{Client: client, ID: id, Image: image, Name: name}
		} else {
			panic(err)
		}
	}

	return &Container{Client: client, ID: res.ID, Image: image, Name: name}
}

func (container *Container) Start() error {
	if err := container.Client.Cli.ContainerStart(
		container.Client.Ctx,
		container.ID,
		dockerContainer.StartOptions{},
	); err != nil {
		return err
	}

	log.Printf("Container %s started.", container.Name)
	return nil
}

func (container *Container) NetworkConnect(networkID string, ip string) error {
	inspect, err := container.Client.Cli.ContainerInspect(container.Client.Ctx, container.ID)
	if err != nil {
		return err
	}

	if _, exists := inspect.NetworkSettings.Networks["IMvlan"]; exists {
		log.Printf("Container %s already connected to network %s.", container.Name, "IMvlan")
		return nil
	}

	if err := container.Client.Cli.NetworkConnect(container.Client.Ctx, networkID, container.ID, &network.EndpointSettings{
		IPAMConfig: &network.EndpointIPAMConfig{
			IPv4Address: ip,
		},
	}); err != nil {
		return err
	}

	log.Printf("Container %s connected to network %s.", container.Name, "IMvlan")
	return nil
}

func GetIdByName(client client.Client, containerName string) (string, error) {
	containers, err := client.Cli.ContainerList(client.Ctx, dockerContainer.ListOptions{All: true})
	if err != nil {
		return "", err
	}

	for _, container := range containers {
		for _, name := range container.Names {
			if name == "/"+containerName {
				return container.ID, nil
			}
		}
	}

	return "", fmt.Errorf("Container ID for %s not found", containerName)
}
