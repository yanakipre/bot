package container

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
)

type Network struct {
	nw *testcontainers.DockerNetwork
}

func NewNetwork(ctx context.Context) (*Network, error) {
	nw, err := network.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot create network: %w", err)
	}

	return &Network{
		nw: nw,
	}, nil
}

func (n *Network) Name() string {
	return n.nw.Name
}

func (n *Network) Close(ctx context.Context) error {
	return n.nw.Remove(ctx)
}
