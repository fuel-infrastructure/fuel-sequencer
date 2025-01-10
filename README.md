# FuelSequencer
**FuelSequencer** is a blockchain built using Cosmos SDK and CometBFT and created with [Ignite CLI](https://ignite.com/cli).

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
- go
    - https://go.dev/doc/install
    - Preferred version: `1.22`

To run the Sequencer, Sidecar, and an Ethereum node:

```bash
make install run-eth-e2e-containers run-sequencer  # terminal 1
make run-sidecar                                   # terminal 2
make clean                                         # once you're done
```

To run just the Sidecar and an Ethereum node:

```bash
make install run-eth-e2e-containers run-sidecar
make clean # once you're done
```

To generate keys for executing transactions:

```bash
make keys
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

### Running Docker Image

The first time you run the docker container you should run it using:

```bash
# with Sidecar
make run-docker-container \
  ETH_RPC_URL="http://example:8545" \
  ETH_WS_URL="ws://example:8545"

# without Sidecar
make run-docker-container \
  ETH_RPC_URL="http://example:8545" \
  ETH_WS_URL="ws://example:8545" \
  COMMAND="fuelsequencerd start"
```

> Note that the Ethereum node URLs need to be set explicitly since it does not make sense to default to `localhost` in the scope of a Docker container. If your Ethereum node is running on `localhost` you will need to specify your host's IP as the Ethereum address. If that does not work, you might need to reconfigure your host firewall to allow Docker to connect on 8545.

The command above will do the following:

1. Create the docker container.
2. Map the chain data located at `./data/fuelsequencer` by default using docker volumes.
3. Give a name to the container.
4. Expose the necessary ports.
5. Start the container.

If you want to pass an alternative data folder you should run the make command as follows:

```bash
make run-docker-container \
  ETH_RPC_URL="http://example:8545" \
  ETH_WS_URL="ws://example:8545" \
  DATA_FOLDER="/path/to/folder/with/config/data/and/keyring-test"
```

After creating and running the docker container for the first time, you should manage the container as follows:

```bash
# To stop the container
make stop-docker-container
```

```bash
# To restart the container
make start-docker-container
```

```bash
# Run make stop-docker-image first if the container has not been stopped yet.
make stop-docker-container

# To remove the container.
make remove-docker-container
```

**NOTE**: Re-running the container using `make run-docker-container` will fail because docker will find an existing 
container with the same name during the creation process.

To follow the container's logs run the following command:

```bash
make follow-docker-logs
```

### Run With Memory Profiling

#### For Development

Before proceeding, ensure that you have a valid Sequencer chain directory located at `./data/fuelsequencer`. If the Sequencer chain directory does not exist, you need to create it by following these steps:

- Remove all profiling-related code from `cmd/fuelsequencerd/main.go`. The sections to be removed are marked with the comment `// TO REMOVE IF CHAIN DIRECTORY NEEDS TO BE INITIALIZED`.
- Run `make install init`. This will initialize the chain directory at `./data/fuelsequencer`.
- Once the chain directory has been created, restore the profiling code in `cmd/fuelsequencerd/main.go`.
- Run `make install` to install the application again.

Run the Sequencer, Sidecar and Pyroscope as follows:

```bash
make run-pyroscope # Starts the Pyroscope server in Docker.
make run-sequencer-with-pyroscope #  Runs the Sequencer with compatibility for Pyroscope
make run-sidecar-with-pyroscope  # Runs the Sidecar with compatibility for Pyroscope
```

Once executed successfully, the Pyroscope UI will be available at http://localhost:4040 for long-term monitoring.

#### For Production

Before proceeding, ensure that a valid Sequencer chain directory exists at `<NODE-HOME>`. If the directory is missing, use an older binary and follow the instructions in this [guide](https://github.com/fuel-infrastructure/networks/tree/main/seq-testnet-2) to create it.

The next step is to run the Pyroscope server:

```bash
make run-pyroscope
```

To enable profiling for the Sequencer, simply replace the old binary with the binary versioned `seq-testnet-2-with-profiling-tag` and restart the daemon service. For detailed instructions on the suggested service file refer to this [guide](https://github.com/fuel-infrastructure/networks/tree/main/seq-testnet-2).

To enable profiling for the Sidecar, use the new binary version `seq-testnet-2-with-profiling-tag` and update the `ExecStart` command in the recommended service file. For detailed instructions on the suggested service file refer to this [guide](https://github.com/fuel-infrastructure/networks/tree/main/seq-testnet-2).

```bash
# Basically just add ExecStart=SERVICE_TYPE="sidecar" to your existing command
ExecStart=SERVICE_TYPE="sidecar" <HOME>/go/bin/fuelsequencerd start-sidecar \
    --host "0.0.0.0" \
    --sequencer_grpc_url "127.0.0.1:9090" \
    --eth_ws_url "<ETHEREUM_NODE_WS>" \
    --eth_rpc_url "<ETHEREUM_NODE_RPC>" \
    --eth_contract_address "0x0E5CAcD6899a1E2a4B4E6e0c8a1eA7feAD3E25eD"
```

Once executed successfully, the Pyroscope UI will be available at http://<vm-ip>:4040 for long-term monitoring.

### References

- https://github.com/Stride-Labs/stride/blob/main/.dockerignore
- https://github.com/Stride-Labs/stride/blob/main/dockernet/start_network.sh
- https://github.com/osmosis-labs/osmosis/blob/3eccca25dd40ec45c0a295f079fb21da66eeeb6a/Dockerfile
- https://github.com/osmosis-labs/osmosis/blob/3eccca25dd40ec45c0a295f079fb21da66eeeb6a/scripts/makefiles/docker.mk
- https://github.com/osmosis-labs/osmosis/blob/main/scripts/makefiles/docker.mk

## Testing

### Generate mocks

```bash
make mocks
```

### Unit tests

```bash
make test-unit
```

### E2E tests

You will need a Sequencer image and Ethereum deployment image:

> Ensure e2e/fuel-rollup/.npmrc file is set up before running this! \
> It should contain `//registry.npmjs.org/:_authToken=<NPM_TOKEN>`. \
> `<NPM_TOKEN>` is an access token to be obtained from your NPM account.

```bash
make build-all-docker-images
```

Then you can run E2E tests:

```bash
make test-e2e
```

## FAQs

### Ran into issues when building/running the chain or the docker image

Most likely this occurs due to permission issues on Linux and the solution is simply to run the following inside the 
project's root directory:

```bash
chmod -R 777 ./
```
