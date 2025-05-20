package client

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
)

type Client struct {
	Ctx context.Context
	Cli *client.Client
}

// Use local (nil) or remote host ("tcp://remote-host:2375")
func NewClient(host *string) (*Client, error) {
	ctx := context.Background()

	// Use remote docker engine
	if host != nil {
		cli, err := client.NewClientWithOpts(client.WithHost(*host), client.WithAPIVersionNegotiation())
		if err != nil {
			return nil, fmt.Errorf("Failed to create client: %w", err)
		}
		return &Client{ctx, cli}, nil
	} else {
		// Use local docker engine
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			return nil, fmt.Errorf("Failed to create client: %w", err)
		}
		return &Client{ctx, cli}, nil
	}
}

func (client *Client) Close() error {
	if client.Cli != nil {
		return client.Cli.Close()
	}
	return nil
}
