# How to contribute

Pull requests are welcomed.

## Review

This repository uses [CodeAssigner](https://github.com/apps/codeassigner) app for the reviews.
Add yourself to http://62.84.113.17/ to get notified about changes.

## Local environment

This describes how to run the repo locally.

### Prerequisites

This repository is using [.tool-versions](https://asdf-vm.com/manage/configuration.html). [Mise](https://github.com/jdx/mise) is recommended for managing the versions of the tools.
Integrate it with your shell and you should have all the needed tools installed.
By cd in the root of the repository, mise will download and install the needed versions of the tools, except for docker-compose and docker. 

Why docker-compose and docker are not included in .tool-versions? They are heavy for the system, many options to have
them exists (podman, orbstack, colima, docker, etc, etc). You need to configure them yourself.

This uses:

* [tilt-dev](https://tilt.dev/)
* docker-compose

### Run the local setup

The following command must be enough to run minimal setup:

```bash
tilt up
```

If you see warnings in the console about missing environment keys,
that means your setup is not complete. You need to provide the missing keys.
