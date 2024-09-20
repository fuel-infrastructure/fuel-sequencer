# gRPC Proxy Demo

- Add `127.0.0.1 grpc.seq.example.com` as a new line in `/etc/hosts/`.
- Run `make init` from the project directory to initialise the Sequencer `data/` folder.
- Run `docker-compose up -d` from the gRPC proxy directory to run the proxy.
- Run any of these example commands to interact with the Sequencer directly:
  - `grpcurl -plaintext localhost:9090 list`
  - `grpcurl -plaintext localhost:9090 describe fuelsequencer.bridge.v1.Query`
  - `grpcurl -plaintext localhost:9090 fuelsequencer.bridge.v1.Query.Params`
- Run any of these example commands to interact with the Sequencer through the proxy:
  - `grpcurl -plaintext grpc.seq.example.com:80 list`
  - `grpcurl -plaintext grpc.seq.example.com:80 describe fuelsequencer.bridge.v1.Query`
  - `grpcurl -plaintext grpc.seq.example.com:80 fuelsequencer.bridge.v1.Query.Params`
