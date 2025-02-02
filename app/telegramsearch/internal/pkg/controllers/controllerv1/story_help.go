package controllerv1

import (
	"context"
)

func (c *Ctl) Help(_ context.Context) string {
	return c.cfg.HelpText
}
