# Manual E2E Testing Scripts

**Requirements**:

- `pip install docker` if you're going to run the scripts against a local testnet running on Docker
- Python (ideally 3.10+).
- If you want to run the tests against the internal or public testnet
  - A FuelSequencer node
  - `fuelsequencerd` that matches the FuelSequencer node version

**Configure E2E tests**:

The following values need to be changed in `run_e2e_tests_X.py` and `run_events_extractor.py`:

- `host`: FuelSequencer IP (e.g. `"localhost"` or a specific IP)
- `port`: FuelSequencer RPC port (e.g. `"26657"`)
- `chain_id`: FuelSequencer chain ID (e.g. `"fuelsequencer"`)
- `binary`: FuelSequencer binary (e.g. `"fuelsequencerd"`)

**Run E2E testing**:

The best way to run the manual E2E testing scripts is to first run python on its own (`python3` in a terminal from the `e2e/manual/` folder), and then start copy-pasting commands from `run_e2e_tests_X.py` into the Python 3 terminal manually. Pay special attention to comments in the E2E testing script as they might indicate that you need to apply some changes (e.g. configuring the connection and channel IDs).

The manual E2E testing framework also comes with an event extractor which can be run directly using `python run_events_extractor.py`.

**Random tips**:

- If a proposal does not pass and you want to know why it didn't, run the following on one of the EntryPoint nodes: `sudo journalctl --reverse -u fuelsequencerd | grep "proposal tallied"`. This will go through logs in reverse chronological order and look up proposal passes/fails.
- If transactions are not going through, consider disabling `wait_for_txs` (e.g. `SEQ.wait_for_txs = False`) so that you can see error messages.
