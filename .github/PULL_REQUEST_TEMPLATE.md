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

Testing and docs:

- [ ] `make proto-routine` for formatting and APIs.
- [ ] `make lint` to ensure linting rules satisfied.
- [ ] `make mocks test-unit` to ensure tests pass with updated mocks.
- [ ] Run E2E tests:
   1. `make build-all-docker-images test-e2e`
   2. `make clean` once you're done.
- [ ] Run a local E2E setup to ensure the chain runs:
   1. `make build-eth-docker-image` to build the latest Ethereum image.
   2. Terminal 1: `make install run-eth-docker-container run-sequencer`
   3. Terminal 2: `make run-sidecar`
   4. Terminal 3: `bash utils/call_contract.sh`
   5. `make clean` once you're done.
