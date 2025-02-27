package container

import (
	"errors"
	"fmt"
	"log"
	"sync"

	dockerContainer "github.com/docker/docker/api/types/container"
	dockerNetwork "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/errdefs"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	"deniable-im/im-sim/internal/types"
	"deniable-im/im-sim/internal/utils/ipv4"
	"deniable-im/im-sim/pkg/client"
	"deniable-im/im-sim/pkg/network"
)

var ErrNoNetworks = errors.New("container not connected to any network")

type Options struct {
	Network         *network.Network
	Ipv4            *string
	ContainerConfig *dockerContainer.Config
	HostConfig      *dockerContainer.HostConfig
	NetworkConfig   *dockerNetwork.NetworkingConfig
	Platform        *v1.Platform
}

type Container struct {
	Client  *client.Client
	ID      string
	Image   string
	Name    string
	Options *Options
}

func NewContainer(client *client.Client, image string, name string, options *Options) (*Container, error) {
	if options == nil {
		options = &Options{
			Network:         nil,
			Ipv4:            nil,
			ContainerConfig: &dockerContainer.Config{},
			HostConfig:      &dockerContainer.HostConfig{},
			NetworkConfig:   &dockerNetwork.NetworkingConfig{},
			Platform:        &v1.Platform{},
		}
	}
	if options.ContainerConfig == nil {
		options.ContainerConfig = &dockerContainer.Config{}
	}
	if options.HostConfig == nil {
		options.HostConfig = &dockerContainer.HostConfig{}
	}
	if options.NetworkConfig == nil {
		options.NetworkConfig = &dockerNetwork.NetworkingConfig{}
	}
	if options.Platform == nil {
		options.Platform = &v1.Platform{}
	}
	if options.Network != nil && options.Ipv4 != nil {
		options.NetworkConfig.EndpointsConfig = map[string]*dockerNetwork.EndpointSettings{
			options.Network.Name: {
				IPAddress: *options.Ipv4,
			},
		}
	}

	options.ContainerConfig.Image = image

	res, err := client.Cli.ContainerCreate(
		client.Ctx,
		options.ContainerConfig,
		options.HostConfig,
		options.NetworkConfig,
		options.Platform,
		name)
	if err != nil {
		if errors.Is(err, errdefs.Conflict(err)) {
			log.Printf("Container %s is already in use.", name)
			id, err := GetIdByName(client, name)
			if err != nil {
				return nil, fmt.Errorf("Failed to create container %w", err)
			}
			return &Container{Client: client, ID: id, Image: image, Name: name, Options: options}, nil
		} else {
			return nil, fmt.Errorf("Failed to create container %w", err)
		}
	}

	return &Container{Client: client, ID: res.ID, Image: image, Name: name, Options: options}, nil
}

func NewContainerSlice(client *client.Client, images []types.Pair[string, string], options *Options) ([]*Container, error) {
	const poolSize = 50
	results := make(chan *Container, poolSize)
	var wg sync.WaitGroup
	errc := make(chan error, len(images))

	for _, image := range images {
		wg.Add(1)

		go func(image types.Pair[string, string]) {
			defer wg.Done()

			// Copy options
			options := *options

			container, err := NewContainer(client, image.Fst, image.Snd, &options)
			if err != nil {
				errc <- fmt.Errorf("Failed to create contailer slice %w", err)
				return
			}

			results <- container
			errc <- nil
		}(image)
	}

	go func() {
		wg.Wait()
		close(results)
		close(errc)
	}()

	var containers []*Container
	for container := range results {
		containers = append(containers, container)
	}

	for err := range errc {
		if err != nil {
			return nil, err
		}
	}

	return containers, nil
}

func (container *Container) Start() error {
	err := container.PruneRedundantNetworks()
	if err != nil && !errors.Is(err, ErrNoNetworks) {
		return fmt.Errorf("Container start prune failed: %w", err)
	}

	if err := container.Client.Cli.ContainerStart(
		container.Client.Ctx,
		container.ID,
		dockerContainer.StartOptions{},
	); err != nil {
		return fmt.Errorf("Container start failed: %w", err)
	}

	opt := container.Options
	if opt != nil && opt.Network != nil && opt.Ipv4 != nil {
		if err := container.NetworkConnect(opt.Network.ID); err != nil {
			return fmt.Errorf("Container start network connect failed: %w", err)
		}
	}

	log.Printf("Container %s started.", container.Name)
	return nil
}

func (container *Container) NetworkConnect(networkID string) error {
	_, err := container.Client.Cli.ContainerInspect(container.Client.Ctx, container.ID)
	if err != nil {
		return fmt.Errorf("Container network connect inspect failed: %w.", err)
	}

	err = container.PruneRedundantNetworks()
	if err != nil && !errors.Is(err, ErrNoNetworks) {
		return fmt.Errorf("Container network connect prune failed: %w", err)
	}

	if err := container.Client.Cli.NetworkConnect(container.Client.Ctx, networkID, container.ID, &dockerNetwork.EndpointSettings{
		IPAMConfig: &dockerNetwork.EndpointIPAMConfig{
			IPv4Address: *container.Options.Ipv4,
		},
	}); err != nil {
		return fmt.Errorf("Container network connect failed: %w", err)
	}

	log.Printf("Container %s connected to network %s.", container.Name, networkID)
	return nil
}

func (container *Container) GetNetworks() ([]types.Pair[string, string], error) {
	inspect, err := container.Client.Cli.ContainerInspect(container.Client.Ctx, container.ID)
	if err != nil {
		return nil, fmt.Errorf("Container get networks %w", err)
	}

	var res []types.Pair[string, string]
	for name, network := range inspect.NetworkSettings.Networks {
		id := network.NetworkID
		res = append(res, types.Pair[string, string]{Fst: name, Snd: id})
	}

	if len(res) == 0 {
		return nil, ErrNoNetworks
	}

	return res, nil
}

func (container *Container) PruneRedundantNetworks() error {
	networks, err := container.GetNetworks()
	if err != nil {
		return fmt.Errorf("Container failed to prune networks %w", err)
	}

	for _, network := range networks {
		networkName := network.Fst
		if err := container.Client.Cli.NetworkDisconnect(container.Client.Ctx, networkName, container.ID, true); err != nil {
			container.Client.Cli.NetworkRemove(container.Client.Ctx, networkName)
			log.Printf("Pruned network %s from container %s.", networkName, container.Name)
		}
	}

	return nil
}

func GetIdByName(client *client.Client, containerName string) (string, error) {
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

func AssignIP(containers []*Container, reservedIP []string, options network.Options) ([]*Container, error) {
	if len(options.IPAM.Config) != 1 {
		return nil, fmt.Errorf("Network assign ip multiple IPAM configs")
	}

	gateway := options.IPAM.Config[0].Gateway
	IPRange := options.IPAM.Config[0].IPRange

	addressSet, err := ipv4.IPv4AddressSpace(IPRange)
	if err != nil {
		return nil, fmt.Errorf("Network assign ip address space: %w", err)
	}

	if len(containers) > len(addressSet) {
		return nil, fmt.Errorf("Network assign ip address space of %d is too low to fit %d containers.", len(addressSet), len(containers))
	}

	delete(addressSet, gateway)
	for _, ip := range reservedIP {
		delete(addressSet, ip)
	}

	for _, container := range containers {
		ip := addressSet.GetFirst()
		container.Options.Ipv4 = &ip
	}

	return containers, nil
}
