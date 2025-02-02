package container

import "slices"

type Option func(c *container)

func WithNetworks(nets ...string) Option {
	return func(c *container) {
		nets = slices.DeleteFunc(nets, func(s string) bool {
			return s == ""
		})
		if len(nets) > 0 {
			c.networks = nets
		}
	}
}

func WithImage(sourceImage string) Option {
	return func(c *container) {
		c.sourceImage = sourceImage
	}
}

func WithHostname(hostname string) Option {
	return func(c *container) {
		c.hostName = hostname
	}
}
