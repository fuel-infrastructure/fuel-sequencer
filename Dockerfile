# syntax=docker/dockerfile:1

# Definition of arg variables.
# These can be overridden at build time with --build-arg
ARG GO_VERSION="1.24.13"
ARG ALPINE_VERSION="3.22"
ARG RUNNER_IMAGE="alpine:${ALPINE_VERSION}"

# --------------------------------------------------------
# Builder
# --------------------------------------------------------

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder

# BuildKit sets these per-platform. Native multi-arch CI builds each arch on a
# matching runner so CGO/ledger compiles correctly (QEMU cross-builds of this
# image historically produced amd64 bits under an arm64 manifest).
ARG TARGETARCH
ARG TARGETOS=linux

# Set the working directory inside the container.
WORKDIR /fuel-sequencer

# Copy the host's package files to the container's workspace.
COPY . /fuel-sequencer

# Install important system dependencies.
RUN apk add --no-cache make git gcc musl-dev openssl-dev linux-headers file

# Download go dependencies.
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/go/pkg/mod \
    go mod download

# Build the fuelsequencerd binary for the target platform.
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/go/pkg/mod \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} make build \
    && test -x /fuel-sequencer/build/fuelsequencerd \
    && file /fuel-sequencer/build/fuelsequencerd

# --------------------------------------------------------
# Runner
# --------------------------------------------------------

FROM ${RUNNER_IMAGE}

# Install some packages and create a fuelsequencer user
RUN apk add bash vim sudo dasel \
    && addgroup -g 1000 fuelsequencer \
    && adduser -S -h /home/fuelsequencer -D fuelsequencer -u 1000 -G fuelsequencer

# Configure the sudoers file by adding fuelsequencer to it
RUN mkdir -p /etc/sudoers.d \
    && echo '%wheel ALL=(ALL) ALL' > /etc/sudoers.d/wheel \
    && echo "%wheel ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers \
    && adduser fuelsequencer wheel

# Get the binary from the previous stage and add it to /usr/local/bin/fu
COPY --from=builder /fuel-sequencer/build/fuelsequencerd /usr/local/bin/fuelsequencerd
# Copy the bash script from the builder
COPY --from=builder /fuel-sequencer/scripts/node_and_sidecar.sh /usr/local/bin/node_and_sidecar

# Set home directory to /home/fuelsequencer
USER 1000
ENV HOME=/home/fuelsequencer
WORKDIR $HOME

# Expose chain ports
EXPOSE 26656
EXPOSE 26657
EXPOSE 1317
EXPOSE 9090

# Expose sidecar ports
EXPOSE 8080
EXPOSE 8081

# Run the script when the container launches
CMD ["node_and_sidecar"]
