package container

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultZookeeperHostname = "zookeeper"
)

type Zookeeper struct {
	container
}

func NewZookeeper(opts ...Option) *Zookeeper {
	c := Zookeeper{
		container: newContainer(defaultZookeeperHostname),
	}

	for _, opt := range opts {
		opt(&c.container)
	}

	return &c
}

func (c *Zookeeper) Run(ctx context.Context) error {
	zk, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "zookeeper:3.7",
			Hostname:     c.container.hostName,
			ExposedPorts: []string{"2181/tcp"},
			WaitingFor:   wait.ForLog("ZooKeeper audit is disabled."),
			Networks:     c.container.networks,
		},
		Started: true,
	})
	if err != nil {
		return fmt.Errorf("could not start clickhouse container: %w", err)
	}

	c.container.container = zk

	return nil
}
