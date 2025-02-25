# Use BuildKit’s Dockerfile frontend to cache local builds.
# https://www.docker.com/blog/containerize-your-go-developer-environment-part-2/
#
# Run docker with DOCKER_BUILDKIT=1 prepended to docker command to activate Build Kit backend.
#
# On linux
# cat /etc/docker/daemon.json
#{ "features": { "buildkit": true } }
# syntax = docker/dockerfile:1-experimental


#
# Telegramsearch app Docker image
#

#
# Intermediate image to build golang API server
#

# Image version version is sha256 to speed up builds (especially local ones) and keep builds reproducible.
# 1.23.2-bullseye
FROM golang@sha256:ecb3fe70e1fd6cef4c5c74246a7525c3b7d59c48ea0589bbb0e57b1b37321fb9 AS go-build
ARG DEBUG=false

# define parent caching directory
ENV CACHE=/cache

# define go caching directories
ENV GOCACHE=$CACHE/gocache
ENV GOMODCACHE=$CACHE/gomodcache

WORKDIR /build

# cache dependencies before building
RUN --mount=type=cache,target=$CACHE \
    --mount=type=bind,source=go.mod,target=go.mod \
    --mount=type=bind,source=go.sum,target=go.sum \
    go mod download -x

RUN mkdir /debugbin; \
    if [ "$DEBUG" = "true" ]; then \
    GOBIN=/debugbin/ go install github.com/go-delve/delve/cmd/dlv@v1.23.1; \
    fi

RUN apt-get update && apt-get install -y git

# build executable, store in cache so rebuilds can no-op
RUN --mount=type=cache,target=$CACHE \
    --mount=type=bind,target=. \
    if [ "$DEBUG" = "true" ]; then \
      go build -buildvcs=auto -o $CACHE/api -gcflags="all=-N -l" ./app/example-app/cmd/api && cp $CACHE/api /bin; \
    else \
      go build -buildvcs=auto -o $CACHE/api ./app/example-app/cmd/api && cp $CACHE/api /bin; \
    fi

# IIUC, there is no such thing as release build in Go. It bundles
# debug symbols by default. We can strip them with `go build -ldflags "-s -w"`,
# but why may want that?

#
# Final image to be exported
#
# Image version is fixed to speed up builds (especially local ones) and keep builds reproducible.
# bullseye-slim
FROM debian@sha256:60a596681410bd31a48e5975806a24cd78328f3fd6b9ee5bc64dca6d46a51f29

# it's an antipattern to run the next lines in multiple steps,
# but we're having frequent 4294967295 exit code and this is an attempt to debug it.
# after it's solved - combine the steps into single one.
RUN apt-get update || cat /var/log/apt/*.log
RUN apt-get install --yes --no-install-recommends \
    ca-certificates \
    openssl || cat /var/log/apt/*.log
RUN apt-get clean || cat /var/log/apt/*.log

RUN useradd -d /example everypay
USER everypay
WORKDIR /example

# Only copy if the /debugbin/dlv exists, otherwise it will do nothing
COPY --chown=everypay --from=go-build /debugbin/dlv* /usr/local/bin/
COPY --from=go-build /bin/api /example/api

EXPOSE 3000 9090
ENTRYPOINT ["/example/api"]
CMD [""]
