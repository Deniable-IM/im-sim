package container

import (
	"fmt"

	dockerContainer "github.com/docker/docker/api/types/container"
	dockerNetwork "github.com/docker/docker/api/types/network"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	"deniable-im/im-sim/internal/logger"
	"deniable-im/im-sim/pkg/network"
)

type Options struct {
	Connections     map[string]*network.Connection
	ContainerConfig *dockerContainer.Config
	HostConfig      *dockerContainer.HostConfig
	NetworkConfig   *dockerNetwork.NetworkingConfig
	Platform        *v1.Platform
}

func NewOptions() *Options {
	return &Options{
		Connections:     nil,
		ContainerConfig: &dockerContainer.Config{},
		HostConfig:      &dockerContainer.HostConfig{},
		NetworkConfig:   &dockerNetwork.NetworkingConfig{},
		Platform:        &v1.Platform{},
	}
}

func (options *Options) SetConfigs(image string, name string) {
	if options.ContainerConfig == nil {
		options.ContainerConfig = &dockerContainer.Config{}
		options.ContainerConfig.Image = image
	} else {
		logger.LogContainerOptions(fmt.Sprintf("[+] ContainerConfig explicit set to overwrite image in container: %s", image, name))
	}

	if options.HostConfig == nil {
		options.HostConfig = &dockerContainer.HostConfig{}
	}

	if options.NetworkConfig == nil {
		options.NetworkConfig = &dockerNetwork.NetworkingConfig{}
		options.NetworkConfig.EndpointsConfig = make(map[string]*dockerNetwork.EndpointSettings)

		for networkName, conn := range options.Connections {
			endPointSettings := &dockerNetwork.EndpointSettings{}
			endPointSettings.IPAMConfig = &dockerNetwork.EndpointIPAMConfig{}
			// Set custom IPv4 address
			if conn.IPv4 != nil {
				endPointSettings.IPAMConfig.IPv4Address = *conn.IPv4
			}

			options.NetworkConfig.EndpointsConfig[networkName] = endPointSettings
		}
	} else {
		logger.LogContainerOptions(fmt.Sprintf("[+] NetworkConfig explicit set to overwrite in container %s", name))
	}

	if options.Platform == nil {
		options.Platform = &v1.Platform{}
	}
}

func (options *Options) DeepCopy() *Options {
	newOptions := &Options{}

	newOptions.Connections = make(map[string]*network.Connection)
	for name, conn := range options.Connections {
		newConn := *conn
		newOptions.Connections[name] = &newConn
	}

	if options.ContainerConfig != nil {
		newOptions.ContainerConfig = &dockerContainer.Config{}
		*newOptions.ContainerConfig = *options.ContainerConfig
	}
	if options.HostConfig != nil {
		newOptions.HostConfig = &dockerContainer.HostConfig{}
		*newOptions.HostConfig = *options.HostConfig
	}
	if options.NetworkConfig != nil {
		newOptions.NetworkConfig = &dockerNetwork.NetworkingConfig{}
		*newOptions.NetworkConfig = *options.NetworkConfig
		if options.NetworkConfig.EndpointsConfig != nil {
			newOptions.NetworkConfig.EndpointsConfig = make(map[string]*dockerNetwork.EndpointSettings)
			for networkName, endPointSettings := range options.NetworkConfig.EndpointsConfig {
				newSettings := *endPointSettings
				newOptions.NetworkConfig.EndpointsConfig[networkName] = &newSettings
			}
		}
	}
	if options.Platform != nil {
		newOptions.Platform = &v1.Platform{}
		*newOptions.Platform = *options.Platform
	}

	return newOptions
}
