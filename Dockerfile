FROM golang:latest as build

WORKDIR /src

# Copy the local package files to the container's workspace
COPY . .

# Build the Go application inside the container
RUN cd /src && go build -o go-infra

FROM alpine:latest as final

# Install any runtime dependencies that are needed to run your application
RUN --mount=type=cache,target=/var/cache/apk \
    apk --update add certbot \
    certbot-dns-cloudflare

WORKDIR /app

COPY --from=build /src/go-infra /app/

# Define the command to keep the container running
ENTRYPOINT ["tail", "-f", "/dev/null"]
