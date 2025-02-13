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

func NewClient() (*Client, error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("Failed to create client: %w", err)
	}
	return &Client{ctx, cli}, nil
}

func (client Client) Close() error {
	if client.Cli != nil {
		return client.Cli.Close()
	}
	return nil
}
