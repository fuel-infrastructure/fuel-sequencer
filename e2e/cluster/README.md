# Cluster Package Documentation

- [Overview](#overview)
  - [Key Features](#key-features)
  - [Important Notes](#important-notes)
- [Running the Cluster](#running-the-cluster)
- [Directory Structure](#directory-structure)
  - [File Overview](#file-overview)
    - [parameters.go](#parametersgo)
    - [setup.go](#setupgo)
    - [build.go](#buildgo)
    - [configure.go](#configurego)
    - [manage.go](#managego)
    - [deploy.go](#deploygo)
    - [connection.go](#connectiongo)
    - [execute.go](#executego)


## Overview

The cluster package provides functionality for setting up and configuring a distributed test network of Fuel Sequencer nodes for end-to-end testing.

### Key Features

1. **Multi-node Support**: Supports setting up multiple validator nodes in a distributed environment
2. **Configuration Management**: Easy configuration through `parameters.go`
3. **Automated Deployment**: Handles binary building, configuration, and deployment automatically
4. **Remote Management**: Supports remote node management through SSH

### Important Notes

- Ensure all validator hosts are accessible via SSH
- Verify network connectivity between validator nodes
- Make sure the user has sufficient permissions on remote systems

This package is primarily used for testing and development purposes. It is not intended for production deployments.

## Running the Cluster

1. Configure parameters in `cluster.toml`

The cluster configuration is now managed through a TOML file located at `e2e/cluster/cluster.toml`. This file contains all the parameters needed for cluster deployment and management.

```toml
[binary]
# Absolute path to the directory where makefile is located
makefile_dir = "/path/to/fuel-sequencer"
# Architecture specified from build binary suffix
want_arch = "linux-amd64"
# Name of the binary on remote
remote_binary_name = "fuelsequencerd"
# Path to the systemd service file on the remote machine
systemd_path = "/etc/systemd/system/fuelsequencerd.service"

[blob]
# Path to the Redis configuration file (relative to makefile_dir)
redis_path = "e2e/cluster/blob/redis.conf"
# Path to the docker compose file for blobpool (relative to makefile_dir)
pool_compose_path = "e2e/cluster/blob/docker-compose.blobpool.yml"
# Path to the docker compose file for blobhub (relative to makefile_dir)
hub_compose_path = "e2e/cluster/blob/docker-compose.blobhub.yml"

# Remote System Parameters
[[systems]]
[systems.destination]
peer_ip = "80.64.208.13"
host = "validator-01.example.com"
user = "username"
pass = "password"
dir = "/home/benchmarks"

[systems.options]
sequencer = true
blob_pool = true
blob_hub = false

# Add more systems as needed
```

The main parameters to configure are `makefile_dir` and the `systems` array for validator nodes.

**Version selection (optional):** You can build and deploy a specific version of fuel-sequencer and/or blob-storage using your local clones:

- `sequencer_version` and `blob_storage_version` in `cluster.toml` can be:
  - unset or `"current"` — use the current tree at `makefile_dir` and `blobhub_dir_path` (default)
  - a **branch name** or **commit hash** — the repo is copied to a temp dir, `git checkout <ref>` is run there, build uses that tree, then the temp dir is removed when setup finishes

2. Run the cluster setup:

```bash
cd e2e
go run ./cmd/cluster
```

Monitor the node logs from the systemd service
```sh
sudo journalctl -xefu fuelsequencerd.service
```

## Directory Structure

```
e2e/cluster/
├── data/             # Data directory generated cluster configurations
├── systemd/          # Systemd service template(s)
├── README.md         # Package documentation
├── build.go          # Binary building functionality
├── configure.go      # Network configuration logic
├── deploy.go         # Network deployment functions
├── main.go           # Main entry point
├── manage.go         # Destination management utilities
├── parameters.go     # Configuration parameters
└── setup.go          # Main setup orchestration
```

### File Overview

#### parameters.go
- Contains all configurable parameters
- Defines network settings and chain configuration
- Specifies validator node destinations

#### setup.go
- Orchestrates the entire setup process
- Handles error logging and management
- Coordinates different setup phases

#### build.go
- Handles building of the Fuel Sequencer binary
- Verifies binary existence and architecture
- Manages build artifacts

#### configure.go
- Sets up validator configurations
- Configures network topology
- Manages genesis state and chain parameters

#### manage.go
- Manages remote destinations
- Handles file transfers and permissions
- Configures systemd services

#### deploy.go
- Handles deployment of the Fuel Sequencer binary and systemd service

#### connection.go
- Lower-level connection management
- Defines methods for establishing and managing SSH connections

#### execute.go
- Lower-level execution commands
- Defines methods for locally and remotely executing commands
- Relays outputs to logger
