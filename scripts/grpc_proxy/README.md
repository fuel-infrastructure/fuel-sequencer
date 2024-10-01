# gRPC Proxy Demo

## Running the Demo

- Add `127.0.0.1 grpc.seq.example.com` as a new line in `/etc/hosts/`.
- Run `make init` from the project directory to initialise the Sequencer `data/` folder.
- Run `docker-compose up -d` from the gRPC proxy directory to run the proxy.
  - This requires that you have ports 81 and 8081 available. If not, adjust from `docker-compose.yml`.
- Run any of these example commands to interact with the Sequencer directly:
  - `grpcurl -plaintext localhost:9090 list`
  - `grpcurl -plaintext localhost:9090 describe fuelsequencer.bridge.v1.Query`
  - `grpcurl -plaintext localhost:9090 fuelsequencer.bridge.v1.Query.Params`
- Run any of these example commands to interact with the Sequencer through the proxy:
  - `grpcurl -plaintext grpc.seq.example.com:81 list`
  - `grpcurl -plaintext grpc.seq.example.com:81 describe fuelsequencer.bridge.v1.Query`
  - `grpcurl -plaintext grpc.seq.example.com:81 fuelsequencer.bridge.v1.Query.Params`

If requests through the proxy are not allowed, you might want to check the proxy's logs `docker logs -f grpc_proxy-reverse-proxy-1` for any errors and ensure that your IP is covered by the IP whitelist in the `dynamic_conf.yml` file. The Traefik UI should also be available at http://localhost:8081.

## Production Use

WARNING: this demo configuration **does not use TLS** (Transport Layer Security) for securing communication between services. It is designed for testing and demonstration purposes only, where encrypted communication is not a priority.

If you plan to deploy this configuration in a production environment, it is crucial to update the configuration to enable TLS to ensure secure, encrypted communication. This involves:

- Enabling HTTPS for secure external communication by configuring Traefik to handle TLS certificates, either via Let's Encrypt or by providing your own certificates.
- Updating backend services to use secure gRPC over TLS if necessary (`https` instead of `h2c`).
- Ensuring proper firewall and security settings to restrict access and secure sensitive data in transit.

Refer to the [Traefik documentation](https://doc.traefik.io/traefik/) on TLS for guidance on configuring certificates and secure communication.

Having the Sequencer run in Docker alongside the proxy is also not the typical setup. You will also need to update the configuration to point to where the Sequencer is deployed.
