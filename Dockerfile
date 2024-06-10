# Use an official Golang runtime as a parent image
FROM golang:alpine

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

# Set the working directory inside the container
WORKDIR /app

# Copy the local package files to the container's workspace
COPY . /app

# Build the Go application inside the container
RUN go build -o go-docker-cli

RUN go install

# Define the command to run your application
ENTRYPOINT ["./go-docker-cli"]