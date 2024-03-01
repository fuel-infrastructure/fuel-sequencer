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

Testing and docs:

- [ ] `make proto-routine` for formatting and APIs.
- [ ] `make lint` to ensure linting rules satisfied.
- [ ] `make test-unit` to ensure tests pass.
- [ ] `make run` to ensure chain runs.
