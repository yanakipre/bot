package container

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
)

type container struct {
	sourceImage string
	networks    []string
	container   testcontainers.Container

	hostName string

	exposedAddr Address
}

func newContainer(hostname string) container {
	return container{
		hostName: hostname,
	}
}

func (c *container) ExposedAddr() Address {
	return c.exposedAddr
}

func (c *container) Stop(ctx context.Context) error {
	return c.container.Terminate(ctx)
}

// portEndpoint replaces DockerContainer.PortEndpoint function.
// Looks like DockerContainer.PortEndpoint function can return incorrect address (use ipv6 port instead of ipv4).
func portEndpoint(
	ctx context.Context,
	c testcontainers.Container,
	port nat.Port,
	ipv6 bool,
) (Address, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return Address{}, err
	}

	// Check if host is a dns name or IP address.
	addr := net.ParseIP(host)
	if addr == nil { // Not an IP address? Let's resolve it!
		ips, err := net.LookupIP(host)
		if err != nil {
			return Address{}, fmt.Errorf("error getting ip address: %w", err)
		}

		for _, ip := range ips {
			if isIPv6(ip.String()) == ipv6 {
				addr = ip
				break
			}
		}
	}

	if addr == nil {
		return Address{}, fmt.Errorf("error resolving address: %s, ipv6: %v", host, ipv6)
	}

	ports, err := mappedPorts(ctx, c, port)
	if err != nil {
		return Address{}, fmt.Errorf("cannot found mapped port: %w", err)
	}

	for _, p := range ports {
		if isIPv6(p.HostIP) == ipv6 {
			return NewAddress(addr.String(), p.HostPort, port.Proto()), nil
		}
	}

	return Address{}, fmt.Errorf("port not found, ipv6: %v", ipv6)
}

func mappedPorts(
	ctx context.Context,
	c testcontainers.Container,
	port nat.Port,
) ([]nat.PortBinding, error) {
	ports, err := c.Ports(ctx)
	if err != nil {
		return nil, err
	}

	for k, p := range ports {
		if k.Port() != port.Port() {
			continue
		}
		if port.Proto() != "" && k.Proto() != port.Proto() {
			continue
		}
		if len(p) == 0 {
			continue
		}

		return p, nil
	}

	return nil, errors.New("port not found")
}

func isIPv6(address string) bool {
	return strings.Count(address, ":") >= 2
}
