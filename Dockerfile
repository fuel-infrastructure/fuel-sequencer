# syntax=docker/dockerfile:1
#
# Native multi-arch rebuild of seq-mainnet-1.7 (DEVOPS-1777 / BOXC-3).
# Keeps GO_VERSION from the seq-mainnet-1.7 tag; adds TARGETOS/TARGETARCH and
# arch self-checks so a Heighliner/QEMU-style mismatch (arm64 manifest + x86
# busybox) cannot ship again.

# Definition of arg variables.
# These can be overridden at build time with --build-arg
ARG GO_VERSION="1.23.12"
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
    && file /fuel-sequencer/build/fuelsequencerd \
    && case "${TARGETARCH}" in \
         amd64) file /fuel-sequencer/build/fuelsequencerd | grep -q 'x86-64' ;; \
         arm64) file /fuel-sequencer/build/fuelsequencerd | grep -Eq 'ARM aarch64|aarch64' ;; \
         *) echo "unsupported TARGETARCH=${TARGETARCH}" >&2; exit 1 ;; \
       esac

# --------------------------------------------------------
# Runner
# --------------------------------------------------------

FROM ${RUNNER_IMAGE}

ARG TARGETARCH

# Link GHCR package back to this repo when published with GITHUB_TOKEN.
LABEL org.opencontainers.image.source="https://github.com/fuel-infrastructure/fuel-sequencer"

# Install some packages and create a fuelsequencer user
RUN apk add --no-cache bash vim sudo dasel file \
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

# Fail the build if shell or binary arch does not match the platform.
# Guards against the Heighliner seq-mainnet-1.7 failure mode (DEVOPS-1771):
# arm64 image shipping x86-64 /bin/sh → chain-init "exec format error" on Graviton.
RUN set -euo pipefail \
    && case "${TARGETARCH}" in \
         amd64) \
           file /bin/sh | grep -q 'x86-64' \
           && file /usr/local/bin/fuelsequencerd | grep -q 'x86-64' ;; \
         arm64) \
           file /bin/sh | grep -Eq 'ARM aarch64|aarch64' \
           && file /usr/local/bin/fuelsequencerd | grep -Eq 'ARM aarch64|aarch64' ;; \
         *) echo "unsupported TARGETARCH=${TARGETARCH}" >&2; exit 1 ;; \
       esac \
    && apk del --no-cache file

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
