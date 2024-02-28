## Description

Closes #xxx

## Checklist

<!-- Mark the following with an 'x' once satisfied -->

PR:

- [ ] Use "Draft:" until ready for review
- [ ] PR directed at `main` branch
- [ ] Pull the latest changes from `main` before requesting review
- [ ] Re-reviewed `Files changed`

State and params:

- [ ] Include new state/param in genesis init and export.
- [ ] Include queries for new state/param.
- [ ] Include set/get/getAll for new state.

Messages:

- [ ] Register new messages in `codec.go`

API:

- [ ] Run `ignite chain build` to ensure `api/` folder is updated.

Testing and docs:

- [ ] Wrote or updated tests
- [ ] Wrote or updated docs
- [ ] Ran linter using `make lint`
- [ ] Ran chain using `make run`
