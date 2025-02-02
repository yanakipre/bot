# Package logfmt

Contains a tool to enforce uniform format for all `zap` log field keys

For example formatting see `logfmt_test.go`

## Usage

```sh
    go run "path/to/logfmt.go --paths /repo/root/path ./relative/sub/path" --verbose
```

- `--paths` quoted space-separated list of paths to recursively search for *.go files in.

    Can be absolute path or relative path to the current working directory.

- `--verbose` if set to true, prints out all unique `zap.Type(k, v)` log field keys found.

