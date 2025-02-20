package network

import (
	"deniable-im/im-sim/pkg/client"
	"github.com/docker/docker/api/types/network"
	"log"
)

type Options struct {
	Driver string
	IPAM   *network.IPAM
}

type Network struct {
	client  client.Client
	Name    string
	ID      string
	options Options
}

func NewNetwork(client client.Client, name string, options Options) *Network {
	inspectRes, err := client.Cli.NetworkInspect(client.Ctx, name, network.InspectOptions{})
	if err != nil {
		createRes, err := client.Cli.NetworkCreate(client.Ctx, name, network.CreateOptions{
			Driver: options.Driver,
			IPAM:   options.IPAM,
		})
		if err != nil {
			panic(err)
		} else {
			return &Network{client: client, Name: name, ID: createRes.ID, options: options}
		}
	}

	log.Printf("Network %s already exists", inspectRes.Name)
	return &Network{client: client, Name: inspectRes.Name, ID: inspectRes.ID, options: options}
}
