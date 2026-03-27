#!/usr/bin/env bash
set -euo pipefail

RPC_URL="${RPC_URL:-https://rpc-fuel-seq.simplystaking.xyz/}"
LOOKUP_DIR="${LOOKUP_DIR:-/Users/thaabl/Documents/Research/Fuel/fuel-sequencer/scripts/lookup_validator}"
PER_PAGE="${PER_PAGE:-100}"
EXPLORER_STAKING_URL="${EXPLORER_STAKING_URL:-https://fuel-seq.simplystaking.xyz/fuel-mainnet/staking}"

BLOCKS_FILE="$(mktemp)"
CACHE_FILE="$(mktemp)"
trap 'rm -f "$BLOCKS_FILE" "$CACHE_FILE"' EXIT

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "Error: required command not found: $1" >&2
    exit 1
  }
}

trim() {
  local s="$1"
  s="${s#"${s%%[![:space:]]*}"}"
  s="${s%"${s##*[![:space:]]}"}"
  printf '%s' "$s"
}

normalize_ts() {
  local ts="$1"
  if [[ "$ts" =~ ^(.*\.[0-9]{3})[0-9]*Z$ ]]; then
    printf '%sZ' "${BASH_REMATCH[1]}"
  else
    printf '%s' "$ts"
  fi
}

rpc_call() {
  local payload="$1"
  curl -sS --request POST \
    --url "$RPC_URL" \
    --header 'Content-Type: application/json' \
    --data "$payload"
}

resolve_validator_name() {
  local addr="$1"
  local cached name operator out line

  cached="$(awk -F'\t' -v a="$addr" '$1==a {print $2 "\t" $3; exit}' "$CACHE_FILE" || true)"
  if [[ -n "$cached" ]]; then
    printf '%s' "$cached"
    return
  fi

  out="$(cd "$LOOKUP_DIR" && go run . "$addr" 2>/dev/null || true)"
  name="UNKNOWN"
  operator=""

  while IFS= read -r line; do
    if [[ "$line" =~ Moniker:[[:space:]]*(.*)$ ]]; then
      name="$(trim "${BASH_REMATCH[1]}")"
    fi
    if [[ "$line" =~ Operator[[:space:]]Addr:[[:space:]]*(.*)$ ]]; then
      operator="$(trim "${BASH_REMATCH[1]}")"
    fi
  done <<< "$out"

  printf '%s\t%s\t%s\n' "$addr" "$name" "$operator" >> "$CACHE_FILE"
  printf '%s\t%s' "$name" "$operator"
}

require_cmd curl
require_cmd jq
require_cmd go
[[ -f "$LOOKUP_DIR/main.go" ]] || { echo "Error: lookup script not found at $LOOKUP_DIR/main.go" >&2; exit 1; }

page=1
while :; do
  payload="$(jq -nc \
    --arg q 'slash.burned_coins > 0' \
    --arg page "$page" \
    --arg per_page "$PER_PAGE" \
    '{
      jsonrpc: "2.0",
      id: 1,
      method: "block_search",
      params: {
        query: $q,
        page: $page,
        per_page: $per_page,
        order_by: "desc"
      }
    }'
  )"

  resp="$(rpc_call "$payload")"

  if jq -e '.error' >/dev/null <<< "$resp"; then
    err_data="$(jq -r '.error.data // ""' <<< "$resp")"
    if [[ "$err_data" == *"page should be within"* ]]; then
      break
    fi
    jq -r '.error' <<< "$resp" >&2
    exit 1
  fi

  count="$(jq -r '.result.blocks | length' <<< "$resp")"
  [[ "$count" -eq 0 ]] && break

  jq -r '.result.blocks[] | [.block.header.height, .block.header.time] | @tsv' <<< "$resp" >> "$BLOCKS_FILE"
  page=$((page + 1))
done

printf 'Block Height\tTimestamp\tValidator\tValidator Signer Address\tReason\tBurned Amount (fuel)\tPower Slashed\tValidator URL\n'

while IFS=$'\t' read -r height timestamp; do
  payload="$(jq -nc --arg h "$height" '{
    jsonrpc: "2.0",
    id: 1,
    method: "block_results",
    params: { height: $h }
  }')"

  resp="$(rpc_call "$payload")"

  if jq -e '.error' >/dev/null <<< "$resp"; then
    continue
  fi

  while IFS= read -r event; do
    addr="$(jq -r '.address // ""' <<< "$event")"
    reason="$(jq -r '.reason // ""' <<< "$event")"
    burned="$(jq -r '.burned_coins // ""' <<< "$event")"
    power="$(jq -r '.power // ""' <<< "$event")"

    IFS=$'\t' read -r validator operator <<< "$(resolve_validator_name "$addr")"
    ts="$(normalize_ts "$timestamp")"
    validator_url=""
    if [[ -n "$operator" ]]; then
      validator_url="$EXPLORER_STAKING_URL/$operator"
    fi

    printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n' \
      "$height" "$ts" "$validator" "$addr" "$reason" "$burned" "$power" "$validator_url"
  done < <(
    jq -c '
      (.result.finalize_block_events // [])
      | map(select(.type == "slash"))
      | .[]
      | (.attributes | map({(.key): .value}) | add)
    ' <<< "$resp"
  )
done < "$BLOCKS_FILE"
