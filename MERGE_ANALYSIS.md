# Merge Analysis: `feature/chunked-blob-consensus`

Merge of the incoming branch into `feature/chunked-blob-consensus`.
Merge base: `834502c` (Partially reverts "fix(blob): harden blobpool and
validation against SQLite lock contention").

Current branch: `e0cfae9` (HEAD).
Incoming branch: `bea0092` (MERGE_HEAD).

---

## 1. Mutually Exclusive Changes

Changes that appeared in only one branch and required no conflict resolution.

### Current branch only

**Exponential backoff on blobhub reconnect** (`x/blob/keeper/blobhub.go`
`sync` method):
- Replaced the fixed 5 s `reconnectBackoff` with exponential backoff (2 s →
  60 s cap).
- Moved the sleep from the connection-error path to a unified pre-reconnect
  block, so both connection failures and channel-close reconnects share the
  same backoff.

**e2e/cluster improvements** (files under `e2e/cluster/`):
- Configurable component versions (`resolve_versions.go`).
- Updated blobhub management and build logic.
- These files are untouched by the incoming branch.

### Incoming branch only

**Proto / generated code — DA attestation message types:**
- `proto/fuelsequencer/blob/tx.proto`: new `SubmitDAAttestation` RPC,
  `DAChunkSig`, `MsgSubmitDAAttestation`, `MsgSubmitDAAttestationResponse`.
- `api/fuelsequencer/blob/tx.pulsar.go`, `tx_grpc.pb.go`, `x/blob/types/tx.pb.go`:
  regenerated from proto.

**x/blob types layer:**
- `types/msg_da_attestation.go`: `ValidateBasic()` and `GetSigners()` for
  `MsgSubmitDAAttestation`.
- `types/codec.go`: registered `MsgSubmitDAAttestation` with the interface
  registry.
- `types/errors.go`: four new sentinel errors (`ErrInvalidAttestation`,
  `ErrInsufficientVotingPower`, `ErrUnknownValidator`, `ErrInvalidSignature`).
- `types/expected_keepers.go`: new `StakingKeeper` interface with
  `GetLastValidators`.
- `types/types.go`: new event type and attribute key constants for DA
  attestation.

**x/blob keeper — on-chain DA attestation handler:**
- `keeper/msg_server_submit_da_attestation.go`: verifies Ed25519 attestation
  signatures against the consensus validator set, weighted by voting power
  (55 % threshold).

**x/blob keeper — chunk-mode attestation pipeline** (all new code in the
incoming branch, originally in `blobhub.go`):
- `syncChunks`: reconnection loop for chunk-mode WebSocket streaming.
- `runChunkPipeline`: fan-out architecture — 16 verify workers, batch
  collector (10 ms tick, max 500), 8 HTTP senders.
- `verifyChunk`: downloads chunk via HTTP GET, BLAKE2b hash, Merkle proof
  verification, Ed25519 signing.
- `verifyMerkleProof` / `hashPairBytes`: binary Merkle tree proof replay.
- `chunkHTTPClient`, `chunkBufPool`: tuned HTTP client and buffer pool.

**x/blob keeper — consensus key loading and new setters:**
- `LoadConsensusKeyPair`: loads Ed25519 key pair from
  `priv_validator_key.json`.
- `SetChunkMode`, `SetChunkValidatorIndex`, `SetConsensusKeyPair`,
  `SetStakingKeeper`.

**x/blob config — chunk-mode flags:**
- `config/config.go`: `ChunkMode` and `ChunkValidatorIndex` fields, CLI
  flags, parsing, validation.

**x/blob module — staking keeper wiring:**
- `module/module.go`: `StakingKeeper` added to `ModuleInputs`, wired via
  `k.SetStakingKeeper(in.StakingKeeper)` in `ProvideModule`.

**app.go wiring:**
- Sets `ChunkMode` and `ChunkValidatorIndex` from config.
- Loads consensus key pair and sets it on the keeper.

**validator_sim** (incoming branch only; excluded from this merge):
- `validator_sim/main.go`: standalone CLI for off-consensus blob sync/ack.
- `validator_sim/chunk.go`: chunk-mode attestation pipeline mirroring the
  keeper's logic with inline binary data support.
- `validator_sim/plan.md`: design document / task specification.
- The entire `validator_sim/` directory was dropped from the merge. It is a
  standalone benchmarking tool that does not affect the sequencer and can be
  introduced in a separate branch.

**.gitignore:**
- Added `vendor/` directory.

**New dependencies:**
- `golang.org/x/crypto` promoted from indirect to direct (BLAKE2b).
- `klauspost/cpuid/v2`, `klauspost/reedsolomon` added as indirect (transitive
  from `blob-storage` erasure coding).

---

## 2. Conflicting Changes and Resolution

Five files were modified in both branches:

### 2.1 `x/blob/keeper/blobhub.go`

**Current branch** changed the `sync` method's reconnect logic from fixed 5 s
backoff to exponential backoff (2 s → 60 s).

**Incoming branch** made extensive changes:
- Extended `blobhubClient` struct with chunk-mode fields, Ed25519 key pair.
- Extended `newBlobhubClient` signature and body (key generation, registration
  via `RegisterWithKey`, chunk-mode branch).
- Changed `Register` → `RegisterWithKey`.
- Added ~350 lines of chunk-mode code (`syncChunks`, `runChunkPipeline`,
  `verifyChunk`, Merkle proof, etc.) appended to the same file.

**Resolution:**
- Took all incoming struct/constructor changes.
- Kept current branch's exponential backoff in `sync` — the incoming branch
  did not touch the `sync` method body beyond what was already at the merge
  base, so this applied cleanly.
- Extracted the incoming branch's chunk-mode code (everything from
  `syncChunks` down) into a new file `x/blob/keeper/chunk.go`. The incoming
  branch had it all in `blobhub.go` (618 lines); splitting keeps each file
  focused.

### 2.2 `x/blob/keeper/keeper.go`

**Current branch** did not change this file.

**Incoming branch** added:
- New keeper fields (`chunkMode`, `chunkValidatorIndex`, `privKey`, `pubKey`,
  `stakingKeeper`).
- New setters (`SetChunkMode`, `SetChunkValidatorIndex`, `SetConsensusKeyPair`,
  `SetStakingKeeper`).
- `LoadConsensusKeyPair` as a new standalone function.
- Updated `Initialize` to pass new parameters to `newBlobhubClient`.

**Resolution:**
- Consolidated two separate incoming functions (`DeriveValidatorIDFromConsensusKey`
  — pre-existing — and `LoadConsensusKeyPair` — new) into a single
  `LoadConsensusKeyInfo` that returns `(validatorID, priv, pub, error)`. This
  avoids loading `priv_validator_key.json` twice and eliminates the redundant
  `SetValidatorID` setter. `DeriveValidatorIDFromConsensusKey` is removed.
- Consolidated `SetConsensusKeyPair` + `SetValidatorID` into a single
  `SetConsensusKeyInfo(validatorID, priv, pub)`.
- Kept all other incoming changes as-is.

### 2.3 `go.mod`

**Current branch** bumped `blob-storage` from `…-bd20dcbd1ea0` to
`…-465d51833f73` (pinned remote tag).

**Incoming branch** switched `blob-storage` to a local `../blob-storage`
replace directive, added `golang.org/x/crypto` as direct dependency, and
added `klauspost/cpuid/v2` + `klauspost/reedsolomon` as indirect.

**Resolution:** took incoming branch's local replace directive
(`../blob-storage`), since both repos are being developed in tandem and the
local path reflects the chunk-mode API surface needed by this branch. Kept
the new direct/indirect dependencies. The commented-out remote pin
(`…-cf7575a7600e`) is the incoming branch's reference commit for when the
local override is removed.

### 2.4 `go.sum` / `e2e/go.mod` / `e2e/go.sum`

**Both branches** changed these due to different `blob-storage` version pins.

**Resolution:** took incoming branch's versions, which align with the local
`../blob-storage` replace. `e2e/go.mod` also uses `../../blob-storage`
replace to stay consistent with root `go.mod`.

### 2.5 `x/blob/keeper/blobhub_test.go`

**Current branch** did not change this file.

**Incoming branch** updated `newBlobhubClient` call sites to pass the new
parameters (`false, 0, nil, nil` for chunk mode and keys).

**Resolution:** took incoming changes directly — the test calls now match the
new function signature.

---

## 3. Review of Resolutions

### 3.1 Chunk code extraction into `chunk.go` — correct

Extracting `syncChunks`, `runChunkPipeline`, `verifyChunk`, `verifyMerkleProof`,
`hashPairBytes`, `chunkHTTPClient`, and `chunkBufPool` from `blobhub.go` into
`chunk.go` is a clean split. `blobhub.go` stays focused on connection
management and whole-blob sync; `chunk.go` owns the chunk attestation
pipeline. No logic was changed during the extraction.

### 3.2 Consolidation of key loading — correct

The merge base had `DeriveValidatorIDFromConsensusKey(nodeHome) → string`.
The incoming branch added `LoadConsensusKeyPair(nodeHome) → (priv, pub)`.
Both load `priv_validator_key.json` independently — double I/O for the same
file.

The resolution's `LoadConsensusKeyInfo(nodeHome) → (validatorID, priv, pub)`
does a single load, derives the validator ID from the public key address (same
as before), and extracts the Ed25519 key pair. `app.go` calls it once and
passes all three to `SetConsensusKeyInfo`.

This is strictly better than the incoming branch's two-function approach.

### 3.3 Exponential backoff aligned across `sync` and `syncChunks`

The `sync` method uses the current branch's exponential backoff (2 s → 60 s).
The incoming branch's `syncChunks` originally used a fixed 5 s backoff. During
this merge the same exponential pattern was applied to `syncChunks` in
`chunk.go`: the fixed `reconnectBackoff` constant was replaced with
`initialReconnectBackoff` / `maxReconnectBackoff`, the sleep was moved from
the post-error path to the unified pre-reconnect `if attempt > 1` block, and
the redundant post-stream-close sleep was removed. Both reconnection loops now
behave identically.

### 3.4 Debug logging added to `verifyChunk`

The incoming branch's `verifyChunk` returned `false` on every failure path
without logging. During this merge, `c.logger.Debug(…)` calls were added to
all five failure paths (request creation, HTTP call, non-200 status, body
read, invalid blob key), each logging the error, blob key, and chunk index.

### 3.5 `validator_sim/` excluded from the merge

The incoming branch included a standalone `validator_sim/` directory (CLI
binary for off-consensus blob sync/ack benchmarking). This was dropped
entirely from the merge — it is independent of the sequencer and can be
introduced in a separate branch. No code in the sequencer depends on it.

### 3.6 Consensus key loading failure logged at Warn

The incoming branch logged the failure at `Debug` level. During this merge it
was changed to `logger.Warn(…)` in `app.go`, so a production validator with a
misconfigured key file sees the message at the default log level.

### 3.7 `go.mod` local replace — appropriate for development

Using `../blob-storage` is the correct choice while both repos are being
developed in tandem. The commented-out remote pin preserves the reference for
when development stabilises.

### 3.8 Registration via `RegisterWithKey` — correct

The incoming branch changed `client.Register(ctx, validatorID)` to
`client.RegisterWithKey(ctx, validatorID, pub)`. This is required by
`blob-storage`'s chunk-mode API, which needs the validator's public key for
Ed25519 signature verification. The merge resolution takes this change.

### 3.9 `blobhub_test.go` call-site updates — correct

Tests pass `false, 0, nil, nil` for the new parameters, preserving whole-blob
mode behaviour. No chunk-mode tests exist yet — see §4.2.

---

## 4. Post-Merge Work

Items that were in neither branch but should be added as follow-up work.

### 4.1 Use `dataChan` from `StreamChunks` in keeper's chunk pipeline

The keeper's `syncChunks` (`chunk.go`) calls
`c.client.StreamChunks(ctx, c.chunkValidatorIndex)` but discards the second
return value (the binary data channel):

```go
notifChan, _, err := c.client.StreamChunks(ctx, c.chunkValidatorIndex)
```

Every chunk then requires a separate HTTP GET to `/chunks/{key}/{index}`.
The `../blob-storage` client's `StreamChunks` supports inline binary data
from the WebSocket stream via the second channel, avoiding the extra HTTP
roundtrip. The keeper should use it for consistency with the streaming
protocol.

### 4.2 Tests for chunk attestation logic

No unit tests exist for:
- `verifyMerkleProof` (pure function, trivially testable).
- `SubmitDAAttestation` message handler (mockable via `StakingKeeper`
  interface).
- `LoadConsensusKeyInfo`.
- The chunk pipeline end-to-end.

### 4.3 Blobhub registration resilience

`newBlobhubClient` (`blobhub.go:86-89`) treats `RegisterWithKey` failure as
fatal — it returns an error that propagates through `Initialize` and prevents
app startup. If blobhub is temporarily unreachable at boot, the entire node
fails to start.

Consider making registration best-effort with background retry, or allowing
startup to proceed without chunk-mode attestation when registration fails.

### 4.4 Integrate unused error sentinels into DA attestation handler

`ErrInsufficientVotingPower`, `ErrUnknownValidator`, and
`ErrInvalidSignature` are defined in `types/errors.go` but not yet referenced.
The handler in `msg_server_submit_da_attestation.go` currently soft-skips
invalid attestations with debug logging. Post-merge, integrate these sentinels
for richer error context — either by wrapping them into the debug log messages
or by returning typed errors where appropriate (e.g., returning
`ErrInsufficientVotingPower` when the 55 % threshold is not met instead of
silently returning `Confirmed: false`).

### 4.5 `ValidateBasic` completeness for `MsgSubmitDAAttestation`

- `BlobSize` is not validated (could be 0). Either add a non-zero check or
  document it as optional/informational.
- `ValidatorAddress` format is not validated beyond non-empty (could be any
  string). Invalid strings will simply fail lookup, so this is not a security
  issue, but early validation gives better error messages.
- `ChunkIndex` has no range check.

### 4.6 `signBlobAsync` unbounded goroutines

Each whole-blob ACK spawns a goroutine (`blobhub.go:262-268`). Under high
blob ingestion this creates unbounded in-flight goroutines. A bounded worker
pool (like the 8 senders in chunk mode) would be safer.

### 4.7 Metrics for chunk attestation monitoring

The chunk pipeline has no Prometheus metrics beyond the shared
`IncrementACKSubmissions` / `IncrementACKErrors` counters. Useful gauges:
- Verify worker queue depth (`len(verifyCh)`).
- Attestation batch queue depth (`len(sendCh)`).
- Verified / failed / signed counters (currently only logged periodically via
  atomic counters).

### 4.8 Pin `blob-storage` to a remote version before merging to `main`

Both `go.mod` and `e2e/go.mod` use local `../blob-storage` replaces. Before
merging `feature/chunked-blob-consensus` into `main`, pin to a tagged or
commit-pinned remote version.

### 4.9 Introduce `validator_sim` as a standalone tool

The incoming branch included `validator_sim/` (standalone CLI for
off-consensus blob sync/ack benchmarking). It was excluded from this merge
(§3.5) as it is independent of the sequencer. It can be introduced in a
separate branch, either in this repo or as its own module.
