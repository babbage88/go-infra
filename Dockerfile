# syntax=docker/dockerfile:1

ARG GO_VERSION=1.22.2
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION} AS build
WORKDIR /src

RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download -x

# This is the architecture you’re building for, which is passed in by the builder.
# Placing it here allows the previous steps to be cached across architectures.
ARG TARGETARCH

# Build the application.
# Leverage a cache mount to /go/pkg/mod/ to speed up subsequent builds.
# Leverage a bind mount to the current directory to avoid having to copy the
# source code into the container.
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,target=. \
    CGO_ENABLED=0 GOARCH=$TARGETARCH go build -o /bin/goinfra ./

#Final image, install python3/cloudflare deps    
FROM alpine:latest AS final

# Install any runtime dependencies that are needed to run your application.
# Leverage a cache mount to /var/cache/apk/ to speed up subsequent builds.
RUN --mount=type=cache,target=/var/cache/apk \
    apk --update add \
    ca-certificates \
    certbot\
    certbot-dns-cloudflare\
    tzdata \
    bash \
    coreutils\
    curl \
    && \
    update-ca-certificates

# Copy the executable from the "build" stage.
COPY --from=build /bin/goinfra /bin/

# Expose the port that the application listens on.
EXPOSE 8993

# What the container should run when it is started.
ENTRYPOINT [ "/bin/goinfra" ]
