package controllerv1

import (
	"context"
)

func (c *Ctl) News(_ context.Context) string {
	return c.cfg.NewsText
}
