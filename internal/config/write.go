package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

// Write serializes your config and writes it to the filesystem,
// for the outputted file to serve as an example.
func Write(cfg any, filename string) {
	marshal := genConfig(cfg)
	create, err := os.Create(filename) //nolint:gosec
	if err != nil {
		panic(err)
	}
	err = create.Close()
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(filename, marshal, os.ModeExclusive); err != nil {
		panic(err)
	}
}

func genConfig(cfg any) []byte {
	marshal, err := yaml.Marshal(cfg)
	if err != nil {
		panic(err)
	}
	return marshal
}
