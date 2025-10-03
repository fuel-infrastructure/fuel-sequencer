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

1. Configure parameters in `parameters.go`

```go
// Example configuration
var (
    makefileDir string = "/path/to/fuel-sequencer"
    wantArch    string = "linux-amd64"
    chainName   string = "clusternet-1"
    ...
    
    // Configure destinations array for validator nodes
    destinations = []destination{
        {
            peer_ip: "80.64.208.13",
            host:    "validator-01.example.com",
            user:    "username",
            pass:    "password",
            dir:     "/home/benchmarks",
        },
        // Add more validator nodes as needed
    }
)
```

As noted in `e2e/cluster/parameters.go`, mainly makefileDir and destinations are the parameters to configure.

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
