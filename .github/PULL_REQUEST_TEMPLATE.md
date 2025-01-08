## Description

Closes #xxx

## Checklist

<!-- Mark the following with an 'x' once satisfied -->

PR:

- [ ] Use "Draft:" until ready for review
- [ ] PR directed at `main` branch
- [ ] Pull latest changes from `main`
- [ ] Re-reviewed `Files changed`

State and params:

- [ ] Include new state/param in genesis init and export.
- [ ] Include queries for new state/param.
- [ ] Include set/get/getAll for new state.

Messages:

- [ ] Register new messages in `codec.go`
- [ ] Added new messages to `handlers_test.go`

Queries:

- [ ] Added new queries to `handlers_test.go`

Metrics:

- [ ] `make metrics` if you updated `metrics.go` files

Testing and docs:

- [ ] `make proto-routine` for formatting and APIs.
- [ ] `make lint` to ensure linting rules satisfied.
- [ ] `make mocks test-unit` to ensure tests pass with updated mocks.
- [ ] Run E2E tests:
   1. Ensure `.npmrc` file is set up in `e2e/fuel-rollup/` with `//registry.npmjs.org/:_authToken=<NPM_TOKEN>`. `<NPM_TOKEN>` is an access token to be obtained from your NPM account.
   2. `make build-all-docker-images test-e2e`
   3. `make clean` once you're done.
- [ ] Run a local E2E setup to ensure the chain runs:
   1. Ensure `.npmrc` file is set up in `e2e/fuel-rollup/` with `//registry.npmjs.org/:_authToken=<NPM_TOKEN>`. `<NPM_TOKEN>` is an access token to be obtained from your NPM account.
   2. Terminal 1: `make install run-eth-e2e-containers run-sequencer`
   3. Terminal 2: `make run-sidecar`
   4. Terminal 3:
      - Wait for the Ethereum deployment container to stop.
      - `bash scripts/get_contract_addresses.sh` to confirm contract addresses (especially for `ethereum_proxy_contract_address` in `config.yml`)
      - `bash scripts/call_contract.sh`
      - Wait for the Sequencer to sync the Ethereum blocks containing the contract calls:
        ```
        fuelsequencerd q bridge last-ethereum-block-synced
        ```
      - Sanity checks:
        ```
        fuelsequencerd q bank balances 0xd447066a8ba9cb15a862a0f6de961f27be86fc0a # expect +20
        fuelsequencerd q bank balances 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266 # expect +80 (+100-10-10)
        fuelsequencerd q block-results 100 # expect supply delta event to be reported
        ```
   5. `make clean` once you're done.
