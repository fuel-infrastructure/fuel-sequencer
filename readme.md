# FuelSequencer
**FuelSequencer** is a blockchain built using Cosmos SDK and Tendermint and created with [Ignite CLI](https://ignite.com/cli).

## Get started

Dependencies:
- buf (only if updating proto files)
    - https://docs.buf.build/installation
    - Preferred version: `1.28.1`
- ignite-cli
    - https://github.com/ignite/cli
    - Preferred version: `v28.2.0`
- proto-builder
    - https://ghcr.io/cosmos/proto-builder
    - Preferred version: `0.14.0`

```bash
make run
```

### Configure

Your blockchain in development can be configured with `config.yml`. To learn more, see the [Ignite CLI docs](https://docs.ignite.com).

## Build for Release

Build binary for various architectures:

```bash
make build-with-checksum
```

Verify checksum:

```bash
cd build && sha256sum -c sha256sum.txt
```

To get an indication as to which binary to run:

```bash
echo "$(uname -s)-$(uname -m)"
```

## Linting

### Linting Go Code

```bash
make lint
```

### Formatting Proto files

Dependencies:
- `docker`: https://docs.docker.com/get-docker/

```bash
make proto-format
```

## Update swagger docs

Dependencies:
- `docker`: https://docs.docker.com/get-docker/

```bash
make proto-swagger-gen
```

## Docker

### Build Docker image

Dependencies:
- `docker`: https://docs.docker.com/get-docker/

```bash
make build-docker-image
```

## Running Docker Image

```bash
make run-docker-image
```

The command above will map the chain data located at `./data/fuelsequencer` by default using docker volumes. If you want 
to pass an alternative data folder you should run the make command as follows:

```bash
make run-docker-image DATA_FOLDER="/path/to/folder/with/config/data/and/keyring-test"
```

### References

- https://github.com/Stride-Labs/stride/blob/main/.dockerignore
- https://github.com/Stride-Labs/stride/blob/main/dockernet/start_network.sh
- https://github.com/osmosis-labs/osmosis/blob/3eccca25dd40ec45c0a295f079fb21da66eeeb6a/Dockerfile
- https://github.com/osmosis-labs/osmosis/blob/3eccca25dd40ec45c0a295f079fb21da66eeeb6a/scripts/makefiles/docker.mk
- https://github.com/osmosis-labs/osmosis/blob/main/scripts/makefiles/docker.mk

## Testing

### Unit tests

```bash
make test-unit
```

### E2E tests

E2E tests are still WIP. This is a list of inspirational code that might be useful when building the E2E tests 
framework:

- https://github.com/osmosis-labs/osmosis/blob/main/tests/e2e/configurer/factory.go#L19
- https://github.com/osmosis-labs/osmosis/blob/3eccca25dd40ec45c0a295f079fb21da66eeeb6a/tests/e2e/containers/containers.go#L487

## FAQs

### Ran into issues when building/running the chain or the docker image

Most likely this occurs due to permission issues on Linux and the solution is simply to run the following inside the 
project's root directory:

```bash
chmod -R 777 ./
```

## Learn more

- [Ignite CLI](https://ignite.com/cli)
- [Tutorials](https://docs.ignite.com/guide)
- [Ignite CLI docs](https://docs.ignite.com)
- [Cosmos SDK docs](https://docs.cosmos.network)
- [Developer Chat](https://discord.gg/ignite)
