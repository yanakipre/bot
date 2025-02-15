// Generator is a go program that generates a conversion routine from status.Status to one of the generated types.
package main

import (
	"bytes"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"text/template"
)

//go:embed convert.go.tmpl
var tmpl []byte

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type args struct {
	PackageName          string
	GeneratedPackagePath string
	GeneratedPackageName string
}

func run() error {
	var (
		packageName          string
		generatedPackagePath string
		destination          string
	)

	flag.StringVar(&packageName, "package-name", "", "package name")
	flag.StringVar(&generatedPackagePath, "generated-package", "", "generated package path")
	flag.StringVar(&destination, "destination", "", "destination filename")

	flag.Parse()

	if packageName == "" {
		return errors.New("package name is required")
	}

	if generatedPackagePath == "" {
		return errors.New("generated package path is required")
	}

	if destination == "" {
		return errors.New("destination filename is required")
	}

	tmpl, err := template.New("tmpl").Parse(string(tmpl))
	if err != nil {
		return fmt.Errorf("failed to parse the template: %w", err)
	}

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "tmpl", &args{
		PackageName:          packageName,
		GeneratedPackagePath: generatedPackagePath,
		GeneratedPackageName: path.Base(generatedPackagePath),
	})
	if err != nil {
		return fmt.Errorf("failed to execute the template: %w", err)
	}

	if err := os.WriteFile(destination, buf.Bytes(), 0o600); err != nil {
		return fmt.Errorf("failed to write to destination file: %w", err)
	}

	return nil
}
