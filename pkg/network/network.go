package network

import (
	"deniable-im/im-sim/internal/logger"
	"deniable-im/im-sim/pkg/client"
	"fmt"

	"github.com/docker/docker/api/types/network"
)

type Options struct {
	Driver string
	IPAM   *network.IPAM
}

type Network struct {
	client  *client.Client
	Name    string
	ID      string
	Options Options
}

func NewNetwork(client *client.Client, name string, options Options) *Network {
	inspectRes, err := client.Cli.NetworkInspect(client.Ctx, name, network.InspectOptions{})
	if err != nil {
		createRes, err := client.Cli.NetworkCreate(client.Ctx, name, network.CreateOptions{
			Driver: options.Driver,
			IPAM:   options.IPAM,
		})
		if err != nil {
			panic(err)
		} else {
			logger.LogNetworkNew(fmt.Sprintf("[+] Network %s created", inspectRes.Scope))
			return &Network{client: client, Name: name, ID: createRes.ID, Options: options}
		}
	}

	logger.LogNetworkNew(fmt.Sprintf("[+] Network %s already exists", inspectRes.Name))
	return &Network{client: client, Name: inspectRes.Name, ID: inspectRes.ID, Options: options}
}

func (network *Network) GetConnection() *Connection {
	return &Connection{Network: network, IPv4: nil}
}
