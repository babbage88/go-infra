# syntax=docker/dockerfile:1

ARG GO_VERSION=1.23.0
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION} AS build
WORKDIR /src

# golang dependencies
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download -x

# Target go version
ARG TARGETARCH

# Build the application, using cache mount.

RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,target=. \
    CGO_ENABLED=0 GOARCH=$TARGETARCH go build -o /bin/server ./

# Final stage copy bin and install pre-requisites
FROM alpine:latest AS final

WORKDIR /app

# Install certbot and the required clouflare dns module.
RUN --mount=type=cache,target=/var/cache/apk \
    apk --update add \
    ca-certificates \
    tzdata \
    certbot \
    certbot-dns-cloudflare \
    && \
    update-ca-certificates

ARG UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    appuser

RUN chown appuser:appuser /app/
USER appuser
RUN mkdir .certbot

#--config-dir ~/.certbot/config --logs-dir ~/.certbot/logs --work-dir ~/.certbot/work

# Copy the executable from the "build" stage.
COPY --from=build /bin/server /app/

# Expose the port that the application listens on.
EXPOSE 8080

COPY entrypoint.sh /app/entrypoint.sh
# What the container should run when it is started.
ENTRYPOINT [ "/app/entrypoint.sh" ]
#ENTRYPOINT ["tail", "-f", "/dev/null"]
