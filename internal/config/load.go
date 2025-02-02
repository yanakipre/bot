package config

import (
	"context"
	"errors"
	"io"
	"os/user"
	"path"

	"github.com/heetch/confita"
	"github.com/heetch/confita/backend"
	"github.com/heetch/confita/backend/env"
	"github.com/heetch/confita/backend/file"
)

type Config interface {
	DefaultConfig()
	Validate() error
}

// Load will load config for given application.
//
// Beware of precedence:
//  1. unmarshalTo is empty config, that provides defaults
//  2. Files are loaded if they exist, overwriting defaults, and each other
//     - /etc/appName/filename
//     - ~/.appName/filename
//     - $(CURDIR)/filename
//  3. Values from environment variables are loaded, overwriting values from files.
func Load(
	ctx context.Context,
	cfgFolder string,
	unmarshalTo Config,
	filename string,
	customBackends ...backend.Backend,
) error {
	unmarshalTo.DefaultConfig()
	usr, err := user.Current()
	if err != nil {
		return err
	}
	dir := usr.HomeDir
	var loaders []backend.Backend
	if filename != "" {
		loaders = append(loaders,
			file.NewOptionalBackend(path.Join("/etc", cfgFolder, filename)),
			file.NewOptionalBackend(path.Join(dir, "."+cfgFolder, filename)),
			file.NewOptionalBackend(filename),
		)
	}
	// we load environment variables after we loaded file config.
	loaders = append(loaders,
		env.NewBackend(),
	)
	loaders = append(loaders, customBackends...)
	l := confita.NewLoader(
		loaders...,
	)
	// yaml tag is needed to:
	//	* parse structs as config
	//	* take values from environment variables
	l.Tag = "yaml"

	err = l.Load(ctx, unmarshalTo)
	if err != nil {
		if !errors.Is(err, io.EOF) {
			// we ignore error with empty config file.
			return err
		}
	}

	return unmarshalTo.Validate()
}
