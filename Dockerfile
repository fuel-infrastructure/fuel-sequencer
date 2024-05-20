# syntax=docker/dockerfile:1

# Definition of arg variables.
ARG GO_VERSION="1.21"
ARG RUNNER_IMAGE="alpine:3.19"

# --------------------------------------------------------
# Builder
# --------------------------------------------------------

FROM golang:${GO_VERSION}-alpine3.19 as builder

# Set the working directory inside the container.
WORKDIR /fuel-sequencer

# Copy the host's package files to the container's workspace.
COPY . /fuel-sequencer

# Install important system dependencies.
RUN apk add --no-cache make git gcc musl-dev openssl-dev linux-headers

# Download go dependencies.
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/go/pkg/mod \
    go mod download

# Build the fuelsequencerd binary.
RUN make build

# --------------------------------------------------------
# Runner
# --------------------------------------------------------

FROM ${RUNNER_IMAGE}

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
ENV HOME /home/fuelsequencer
WORKDIR $HOME

# Expose chain ports
EXPOSE 26656
EXPOSE 26657
EXPOSE 1317
EXPOSE 9090

# Expose sidecar ports
EXPOSE 8080

# Run the script when the container launches
CMD ["node_and_sidecar"]
