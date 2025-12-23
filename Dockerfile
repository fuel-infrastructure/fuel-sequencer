# syntax=docker/dockerfile:1

# Definition of arg variables.
ARG GO_VERSION="1.23.11"
ARG RUNNER_VERSION="3.22"

# --------------------------------------------------------
# Builder
# --------------------------------------------------------

FROM golang:${GO_VERSION}-alpine${RUNNER_VERSION} AS builder

# Set the working directory inside the container.
WORKDIR /fuel-sequencer

# Copy the host's package files to the container's workspace.
COPY . /fuel-sequencer

# Install important system dependencies.
RUN apk add --no-cache make git gcc musl-dev openssl-dev linux-headers openssh-client

# Configure Git to use SSH for GitHub
RUN git config --global url."git@github.com:".insteadOf "https://github.com/"

# Add GitHub to known_hosts to avoid host key verification failures
# GitHub host keys: https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/githubs-ssh-key-fingerprints
RUN mkdir -p /root/.ssh && \
    ssh-keyscan github.com >> /root/.ssh/known_hosts 2>/dev/null || \
    (echo "github.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl" >> /root/.ssh/known_hosts)

# Download go dependencies.
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/go/pkg/mod \
    --mount=type=ssh \
    go mod download

# Build the fuelsequencerd binary.
RUN make build

# --------------------------------------------------------
# Runner
# --------------------------------------------------------

FROM alpine:${RUNNER_VERSION}

# Get the binary from the previous stage and add it to /usr/local/bin/fu
COPY --from=builder /fuel-sequencer/build/fuelsequencerd /usr/local/bin/fuelsequencerd
# Copy the bash script from the builder
COPY --from=builder /fuel-sequencer/scripts/node_and_sidecar.sh /usr/local/bin/node_and_sidecar

# Install some packages and create a fuelsequencer user
RUN apk add bash vim sudo dasel \
    && addgroup -g 1000 fuelsequencer \
    && adduser -S -h /home/fuelsequencer -D fuelsequencer -u 1000 -G fuelsequencer

# Configure the sudoers file by adding fuelsequencer to it
RUN mkdir -p /etc/sudoers.d \
    && echo '%wheel ALL=(ALL) ALL' > /etc/sudoers.d/wheel \
    && echo "%wheel ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers \
    && adduser fuelsequencer wheel

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
