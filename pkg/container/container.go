package container

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	dockerContainer "github.com/docker/docker/api/types/container"
	dockerNetwork "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/errdefs"

	"deniable-im/im-sim/internal/logger"
	"deniable-im/im-sim/internal/types"
	"deniable-im/im-sim/internal/utils/ipv4"
	"deniable-im/im-sim/pkg/client"
	"deniable-im/im-sim/pkg/network"
	"deniable-im/im-sim/pkg/process"
)

var ErrNoNetworks = errors.New("container not connected to any network")

type Container struct {
	Client  *client.Client
	ID      string
	Image   string
	Name    string
	Options *Options
}

func NewContainer(client *client.Client, image string, name string, options *Options) (*Container, error) {
	if options == nil {
		options = NewOptions()
	}

	options.SetConfigs(image, name)

	res, err := client.Cli.ContainerCreate(
		client.Ctx,
		options.ContainerConfig,
		options.HostConfig,
		options.NetworkConfig,
		options.Platform,
		name)
	if err != nil {
		if errors.Is(err, errdefs.Conflict(err)) {
			// Container is already in use! Return it
			id, err := GetIdByName(client, name)
			if err != nil {
				return nil, fmt.Errorf("Failed to create container conflict: %w.", err)
			}
			return &Container{Client: client, ID: id, Image: image, Name: name, Options: options}, nil
		} else {
			return nil, fmt.Errorf("Failed to create container: %w.", err)
		}
	}

	return &Container{Client: client, ID: res.ID, Image: image, Name: name, Options: options}, nil
}

func NewContainerSlice(client *client.Client, images []types.Pair[string, string], options *Options) ([]*Container, error) {
	imagesLen := len(images)

	const poolSize = 50
	results := make(chan *Container, poolSize)
	var wg sync.WaitGroup
	errc := make(chan error, len(images))

	reader, writer := io.Pipe()
	go func() {
		logger.LogContainerSlice(reader)
	}()

	for _, image := range images {
		wg.Add(1)

		go func(image types.Pair[string, string]) {
			defer wg.Done()

			options := options.DeepCopy()

			container, err := NewContainer(client, image.Fst, image.Snd, options)
			if err != nil {
				errc <- fmt.Errorf("Failed to create contailer slice: %w.", err)
				return
			}

			results <- container

			log := fmt.Sprintf(
				`{"status": "created", "total": %d, "image": "%s", "name": "%s"}`,
				imagesLen,
				image.Fst,
				image.Snd)
			writer.Write([]byte(log))

			errc <- nil
		}(image)
	}

	go func() {
		wg.Wait()
		close(results)
		close(errc)
		writer.Close()
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

	reader.Close()

	return containers, nil
}

func StartContainers(containers []*Container) {
	containersLen := len(containers)
	const poolSize = 50
	var wg sync.WaitGroup
	errc := make(chan error, len(containers))

	reader, writer := io.Pipe()
	go func() {
		logger.LogStartContainers(reader)
	}()

	for _, container := range containers {
		wg.Add(1)

		go func(c *Container) {
			defer wg.Done()
			if err := c.start(); err != nil {
				errc <- err
			}

			log := fmt.Sprintf(
				`{"status": "started", "total": %d, "image": "%s", "name": "%s"}`,
				containersLen,
				container.Image,
				container.Name)
			writer.Write([]byte(log))
		}(container)
	}

	go func() {
		wg.Wait()
		close(errc)
		writer.Close()
	}()

	for err := range errc {
		if err != nil {
			panic(err)
		}
	}

	reader.Close()
}

func (container *Container) start() error {
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

	connections := container.Options.Connections
	for _, conn := range connections {
		if err := container.NetworkConnect(*conn.Network); err != nil {
			return fmt.Errorf("Container start network connect failed: %w", err)
		}
	}
	return nil
}

func (container *Container) Start() error {
	if err := container.start(); err != nil {
		return err
	}
	logger.LogContainerStarted(fmt.Sprintf("Container %s started.", container.Name))
	return nil
}

func (container *Container) NetworkConnect(network network.Network) error {
	_, err := container.Client.Cli.ContainerInspect(container.Client.Ctx, container.ID)
	if err != nil {
		return fmt.Errorf("Container network connect inspect failed: %w.", err)
	}

	// Update IPv4 if assigned later
	endPointSettings := container.Options.NetworkConfig.EndpointsConfig[network.Name]
	conn := container.Options.Connections[network.Name]
	if endPointSettings != nil && conn != nil && conn.IPv4 != nil {
		endPointSettings.IPAMConfig = &dockerNetwork.EndpointIPAMConfig{}
		endPointSettings.IPAMConfig.IPv4Address = *conn.IPv4
	}

	if err := container.Client.Cli.NetworkConnect(container.Client.Ctx, network.ID, container.ID, endPointSettings); err != nil {
		return fmt.Errorf("Container network connect failed: %w", err)
	}

	logger.LogNetworkConnect(fmt.Sprintf("[+] Container %s connected to network %s", container.Name, network.Name))
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

func AssignIP(containers []*Container, reservedIP []string, net network.Network) ([]*Container, error) {
	ipam := net.Options.IPAM
	if len(ipam.Config) != 1 {
		return nil, fmt.Errorf("Network assign ip need one IPAM config. Got %d.", len(ipam.Config))
	}

	gateway := ipam.Config[0].Gateway
	IPRange := ipam.Config[0].IPRange

	addressSet, err := ipv4.IPv4AddressSpace(IPRange)
	if err != nil {
		return nil, fmt.Errorf("Network assign ip address space: %w.", err)
	}

	if len(containers) > len(addressSet) {
		return nil, fmt.Errorf("Network assign ip address space of %d is too low to fit %d containers.", len(addressSet), len(containers))
	}

	delete(addressSet, gateway)
	for _, ip := range reservedIP {
		delete(addressSet, ip)
	}

	for _, container := range containers {
		ip := addressSet.Pop()
		container.Options.Connections[net.Name].IPv4 = &ip
	}

	return containers, nil
}

// User responsible for Process.Close()
func (container *Container) Exec(commands []string, logOutput bool) (*process.Process, error) {
	options := dockerContainer.ExecOptions{
		Cmd:          commands,
		AttachStdout: true,
		AttachStderr: true,
		AttachStdin:  true,
		Tty:          false,
		Detach:       false,
	}

	execRes, err := container.Client.Cli.ContainerExecCreate(container.Client.Ctx, container.ID, options)
	if err != nil {
		return nil, fmt.Errorf("Container Exec failed to create: %w.", err)
	}

	res, err := container.Client.Cli.ContainerExecAttach(container.Client.Ctx, execRes.ID, dockerContainer.ExecStartOptions{})
	if err != nil {
		return nil, fmt.Errorf("Container Exec failed to attach: %w.", err)
	}
	res.Conn.SetReadDeadline(time.Time{})
	var buffer bytes.Buffer

	tee := io.TeeReader(res.Reader, &buffer)

	if logOutput {
		go logger.LogContainerExec(tee, commands, container.Name)
	} else {
		go func() {
			io.Copy(io.Discard, tee)
		}()
	}

	execFunc := container.Exec
	return process.NewProcess(res.Conn, &buffer, commands, execFunc), nil
}
