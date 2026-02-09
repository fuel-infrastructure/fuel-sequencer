---
name: validator_sim blob sync
overview: "Add a minimal standalone Go binary `validator_sim` that performs the off-consensus x/blob flow: stream blobs (or chunk notifications) from blobhub, sync to blobpool or attest chunks, and ack/sign back to blobhub. Supports both whole-blob and chunk-mode attestation. No MsgBlobMetadataTx; uses blob-storage client and store only."
todos: []
isProject: false
---

# validator_sim: off-consensus blob sync/ack CLI

## Goal

A small Go command that does only the **off-consensus** part of x/blob in one of two modes:

**Whole-blob mode (default):**

1. **Stream** blobs from blobhub (WebSocket via blob-storage client).
2. **Sync** each blob into a local blobpool (sqlite store).
3. **Ack/sign** back to blobhub (one signature per blob key) when running as a validator.

**Chunk-mode** (optional):

1. **Stream** chunk notifications from blobhub for this validator’s assigned chunk index (no full blobs).
2. For each notification: **download** chunk via GET, **verify** hash and Merkle proof, **sign** chunk attestation (blob_key + chunk_index + chunk_hash), **submit** via `SignChunkBatch`. No blobpool for full blobs; ACK is per chunk.

No on-chain or RPC parts: no `MsgBlobMetadataTx`, no blobpool HTTP server.

## Reference behavior (to mirror)

The logic lives in [x/blob/keeper/blobhub.go](x/blob/keeper/blobhub.go) and [x/blob/keeper/blobpool.go](x/blob/keeper/blobpool.go):

- **Sync loop** ([blobhub.go:107–211](x/blob/keeper/blobhub.go)): connect via `client.StreamBlobs(ctx)`, then for each blob: skip if `blobpool.Has(blob.Key)`, else `blobpool.Insert(ctx, blob.Data)`; if `validatorID != ""` call `signBlobAsync(ctx, blob.Key)`.
- **Sign/ack** ([blobhub.go:216–264](x/blob/keeper/blobhub.go)): `client.Sign(ctx, validatorID, key)` with retries (e.g. 3x, 5s backoff); run in a goroutine so sync is not blocked.
- **Blobpool** ([blobpool.go:38–67](x/blob/keeper/blobpool.go)): sqlite store via `sqlite.Store(ctx, path, false)`, optional `store.Prune(ctx, 20*time.Minute)`; `Has`/`Put` map to `store.Has`/`store.Put`.
- **Registration**: if validator ID is set, call `client.RegisterWithKey(ctx, validatorID, pub)` before starting sync ([blobhub.go:89–95](x/blob/keeper/blobhub.go)).

**Chunk-mode** ([blobhub.go:290–353](x/blob/keeper/blobhub.go), [355–485](x/blob/keeper/blobhub.go), [513–581](x/blob/keeper/blobhub.go)): connect via `client.StreamChunks(ctx, chunkValidatorIndex)` to receive `ChunkNotification` (blob key, chunk index, Merkle root, proof). Pipeline: (1) **Verify** each chunk: GET `{blobhub}/chunks/{blobKey}/{chunkIndex}`, compute blake2b chunk hash, verify Merkle proof against root; build attestation message `blob_key(32) || chunk_index(4 BE) || chunk_hash(32)` and Ed25519-sign with validator key. (2) **Batch** attestations (e.g. 10ms ticker, max 500 per batch). (3) **Send** via `client.SignChunkBatch(ctx, validatorID, atts)`. Reconnect on stream close with same backoff. No blobpool; ACK is per chunk.

Implement the same flows in a **standalone** binary that depends only on `github.com/fuel-infrastructure/blob-storage` (client + store/sqlite), so we do not need to export keeper internals and the script stays minimal.

## Placement and deps

- **Location**: [validator_sim/](validator_sim/) at **repo root** (e.g. `validator_sim/main.go` and any `validator_sim/*.go`). Not under e2e; sibling to `app/`, `e2e/`, `x/`, etc.
- **Module**: Root module `github.com/fuel-infrastructure/fuel-sequencer` (use root go.mod). Root already has `github.com/fuel-infrastructure/blob-storage`; no e2e module dependency.
- **Logging**: standard `log` or a simple logger to keep the binary lightweight.

## CLI compatibility with run_benchmark.sh

The script [run_benchmark.sh](file:///Users/thaabl/Documents/Research/Fuel/blob-storage-vitaly/run_benchmark.sh) invokes validator_sim as:

```bash
/tmp/validator_sim validator-1 http://localhost:31035 --chunk-mode --index 0
```

**Do not change the script.** The binary must accept:

- **Positional args (in order):** `<validator-id>` (e.g. `validator-1`), `<blobhub-url>` (e.g. `http://localhost:31035`).
- **Flags:** `--chunk-mode`, `--index N` (chunk validator index, e.g. 0, 1, 2).

So: validator ID and blobhub URL come from positionals when running under the script; chunk mode and index from flags. Optional flags (e.g. `-blobpool`, `-key`) can be added for standalone/whole-blob use without breaking the script.

## Flags and positional args

| Input | Default | Purpose |

|-------|--------|--------|

| **Positional 1** | (required when used as in script) | Validator ID (e.g. `validator-1`). |

| **Positional 2** | (required when used as in script) | Blobhub URL (e.g. `http://localhost:31035`). |

| `--chunk-mode` | false | Chunk-mode attestation: stream chunk notifications, verify and sign chunks only. Script uses this. |

| `--index` | 0 | Chunk index this validator attests (script uses 0, 1, 2 for three validators). |

| `-blobpool` | `./data/blobpool.db` | SQLite path for blob store (whole-blob mode only). Optional. |

| `-key` | (empty) | Path to Ed25519 private key file. If validator ID set and `-key` empty, generate key and log hex. Optional. |

Parsing: when both positionals are provided, use them for validator ID and blobhub URL; otherwise allow `-validator-id` and `-blobhub` as fallbacks for standalone use. Normalize blobhub URL (prepend `http://` if no scheme). When `--chunk-mode` is true, blobpool is not used.

## Implementation similarity to x/blob

Keep the implementation as close as possible to [x/blob/keeper/blobhub.go](x/blob/keeper/blobhub.go) so that logic can be compared or synced easily:

- **Same flow and names:** Reconnect loop with same backoff (5s); whole-blob path mirrors `sync()` (StreamBlobs → Has/Insert → signBlobAsync); chunk path mirrors `syncChunks()` → `runChunkPipeline()` with the same pipeline stages (verify workers → batch collector → send workers).
- **Same constants where applicable:** Chunk pipeline: e.g. 16 verify workers, batch ticker 10ms, max 500 per batch, 8 send workers (see [blobhub.go:384–473](x/blob/keeper/blobhub.go)); reconnect backoff 5s; sign retries 3x, 5s delay.
- **Same verification logic:** Copy or closely mirror `verifyChunk` (GET chunk, blake2b hash, Merkle proof check, attestation message `blob_key(32) || chunk_index(4 BE) || chunk_hash(32)`, Ed25519 sign) and `verifyMerkleProof` / `hashPairBytes` from [blobhub.go:513–607](x/blob/keeper/blobhub.go) so chunk attestations are identical to x/blob.
- **Same client/store usage:** Use blob-storage `Client` (StreamBlobs, StreamChunks, Sign, SignChunkBatch, RegisterWithKey) and store (Has, Put, Prune) so behavior matches keeper.

## Implementation steps

1. **Add** [validator_sim/main.go](validator_sim/main.go) (and optionally other files under `validator_sim/*.go` if the logic is split):

   - **CLI:** Parse positionals first: if two args, use as `<validator-id>` and `<blobhub-url>` (script style); else support `-validator-id` and `-blobhub` for standalone. Parse flags: `--chunk-mode`, `--index` (chunk index), `-blobpool`, `-key`. Set up `context` with cancel (e.g. on OS signal). When `--chunk-mode` is true, validator ID is required (from positional or flag).
   - Create blob client: `blobclient.NewClient(&blobclient.ClientConfig{ BaseURL: normalizedBlobhubURL })`.
   - If validator ID is set: load key from `-key` or generate; call `client.RegisterWithKey(ctx, validatorID, pub)`.
   - **Branch on mode** (mirror [blobhub.go](x/blob/keeper/blobhub.go) structure):
     - **Whole-blob (default):** Create sqlite store at `-blobpool`, `store.Prune(ctx, 20*time.Minute)` if supported. Reconnect loop (same as keeper): `StreamBlobs(ctx)`; for each blob: if `store.Has(ctx, blob.Key)` skip; else `store.Put(ctx, blob.Data)`; if validator ID set, goroutine `client.Sign(ctx, validatorID, blob.Key)` with retries (3x, 5s). Handle nil blob, channel close, context cancel.
     - **Chunk-mode:** No blobpool. Reconnect loop: `StreamChunks(ctx, index)`; run chunk pipeline matching keeper: (a) 16 verify workers (verifyChunk: GET chunk, blake2b, verifyMerkleProof, build ChunkAttestation with same 68-byte signed message); (b) batch collector (10ms ticker, max 500); (c) 8 send workers `SignChunkBatch(ctx, validatorID, batch)`. Same backoff (5s) on stream close.
   - Shared: normalize blobhub URL; 5s backoff; clean exit on context cancel.

2. **Key handling**: Read Ed25519 key from `-key` (hex or raw 32 bytes); if `-validator-id` is set and `-key` is empty, generate with `ed25519.GenerateKey(nil)` and log the hex-encoded private key so the user can save it and reuse.

3. **Chunk verification**: Mirror keeper’s `verifyChunk` and `verifyMerkleProof` / `hashPairBytes` ([blobhub.go:513–607](x/blob/keeper/blobhub.go)) so attestation format and Merkle verification are identical. Use blob-storage client types for `ChunkNotification` and `ChunkAttestation`.

4. **Optional**: Add a short README or comment in main describing both modes (whole-blob: sync → blobpool, ack blob key; chunk-mode: stream chunks → verify → sign chunk attestations) and that MsgBlobMetadataTx is out of scope.

## Out of scope (by design)

- **MsgBlobMetadataTx** and any on-chain or app-level RPC.
- **Blobpool server** (HTTP server for querying blobs).
- **Metrics** (e.g. Prometheus); can be added later.

## Testing

- Manual: (1) Whole-blob: run blobhub locally, run `validator_sim` with blobhub URL and optional validator-id/key (positionals or flags), confirm blobs in sqlite and acks sent. (2) Chunk-mode / script: run as in [run_benchmark.sh](file:///Users/thaabl/Documents/Research/Fuel/blob-storage-vitaly/run_benchmark.sh): `validator_sim validator-1 http://localhost:31035 --chunk-mode --index 0` (and index 1, 2 for two more); confirm chunk attestations submitted. To use from the script, point the script’s build at fuel-sequencer’s `validator_sim` dir (e.g. `(cd /path/to/fuel-sequencer && go build -o /tmp/validator_sim ./validator_sim)`).
- Optional later: unit test with a mock blobhub (e.g. similar to [x/blob/keeper/blobhub_test.go](x/blob/keeper/blobhub_test.go) mock) to assert sync + sign behavior without a real server.

## Build

From repo root:

```bash
go build -o validator_sim ./validator_sim
```

No new Makefile target required unless you want one for convenience.