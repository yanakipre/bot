# Image version version is sha256 to speed up builds (especially local ones) and keep builds reproducible.
# 1.23.2-alpine
FROM golang@sha256:9dd2625a1ff2859b8d8b01d8f7822c0f528942fe56cfe7a1e7c38d3b8d72d679 AS build


WORKDIR /build

COPY go.mod .
COPY go.sum .

# define parent caching directory
ENV CACHE=/cache

# define go caching directories
ENV GOCACHE=$CACHE/gocache
ENV GOMODCACHE=$CACHE/gomodcache

# cache dependencies before building
RUN --mount=type=cache,target=$CACHE \
    --mount=type=bind,source=go.mod,target=go.mod \
    --mount=type=bind,source=go.sum,target=go.sum \
    go mod download -x

RUN --mount=type=cache,target=$CACHE \
    --mount=type=bind,target=. \
    go build -o $CACHE/postgres-migrate ./internal/db-management/cmd/postgres-migrate && cp $CACHE/postgres-migrate /bin/

# Image version is fixed to speed up builds (especially local ones) and keep builds reproducible.
# 3.14.10
FROM alpine@sha256:0f2d5c38dd7a4f4f733e688e3a6733cb5ab1ac6e3cb4603a5dd564e5bfb80eed

RUN apk add --no-cache bash ca-certificates

# https://github.com/moby/moby/issues/30081#issuecomment-1137512170
RUN mkdir -p /db
WORKDIR /db

COPY --from=build /bin/postgres-migrate /usr/local/bin/postgres-migrate
COPY app/telegramsearch/telegramsearch-app/migrations /db/migrations

CMD ["postgres-migrate"]
