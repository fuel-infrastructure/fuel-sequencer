# Cosmos SDK Validator Location Scanner

Determines the geographic location of the top N% of validators (by voting power, default 66.7%) in a Cosmos SDK network using multiple geolocation sources for reliability.

## Why Multiple Sources?

Single IP geolocation sources are often inaccurate, especially for:
- Cloud/datacenter IPs (AWS, GCP, Hetzner, etc.)
- VPN endpoints
- Anycast IPs
- Recently reassigned IP blocks

This tool queries **5 different geolocation services** and cross-references them to calculate a confidence score. Country names are normalized (ISO alpha-2 codes like "US" are mapped to full names like "United States") so votes from different sources unify correctly.

## Tools Provided: Go Version (`validator-locator`)

**Advantages:**
- Concurrent API queries (faster)
- Better error handling
- Cleaner code structure
- Easier to extend

**Build:**
```bash
go build -o validator-locator .
```

**Usage:**
```bash
# Basic usage (top 66.7% by voting power)
./validator-locator

# Custom percentage threshold (e.g. top 50%)
./validator-locator --percentage 50

# With traceroute verification (enabled by default, disable with --traceroute=false)
./validator-locator --traceroute=false

# JSON output
./validator-locator --json

# All validators (100%)
./validator-locator -p 100

# Custom endpoints
RPC_URL=https://rpc.cosmos.network REST_URL=https://rest.cosmos.network ./validator-locator
```

## Geolocation Sources Used

| Source | Rate Limit | Notes |
|--------|------------|-------|
| ip-api.com | 45/min | Good accuracy, includes datacenter detection |
| ipinfo.io | 50k/month | Reliable, includes ASN |
| freeipapi.com | Generous | Returns full country names, no key needed |
| ipwho.is | Unlimited | Includes ISP/org/connection info |
| WHOIS | N/A | Authoritative for ASN/org data, country codes only |
| Reverse DNS | N/A | Parses PTR hostnames for datacenter location codes (e.g. `fsn` → Falkenstein) |
| Traceroute | N/A | Optional, only contributes when route reaches destination |

## How IP Discovery Works

The tool attempts to find validator IPs through multiple methods:

1. **Peer Network Info** (`/net_info` RPC endpoint)
   - Matches validator monikers to connected peer IPs
   - Uses fuzzy matching (exact, first word, prefix)

2. **Website DNS Resolution**
   - If validator has a website, resolves its IP
   - Useful when peer IP isn't available

3. **Traceroute Analysis**
   - Traces route to target IP
   - Only used when the route actually reaches the destination IP
   - If the destination is not found in the traceroute output, the result is discarded to avoid reporting intermediate hop locations

## Country Normalization

Different geolocation sources return country information in different formats:
- `ip-api.com`, `ipwho.is`, `freeipapi.com` return full names (e.g., "United States")
- `ipinfo.io` returns ISO alpha-2 codes (e.g., "US")
- `whois` returns ISO alpha-2 codes (e.g., "US")

The tool normalizes all country values to full names before voting, so "US" and "United States" count as the same vote.

## Confidence Scoring

The confidence score (0-100%) is calculated based on:

- **Country Agreement (40% weight):** How many sources agree on country
- **City Agreement (60% weight):** How many sources agree on city

| Score | Level | Meaning |
|-------|-------|---------|
| 80-100% | High | Strong consensus across sources |
| 50-79% | Medium | Some disagreement, likely accurate country |
| 0-49% | Low | Sources disagree significantly |

## Output Example

```
Rank: 1
  Operator Address: cosmosvaloper1xyz...
  Moniker: ExampleValidator
  Voting Power: 1000000 tokens (5.23%)
  Website: https://example.com
  Peer IP: 203.0.113.1
  Geographic Location: Frankfurt, Hesse, Germany
  Network: AS24940 (Hetzner Online GmbH)
  Datacenter: hosting
  Confidence: 85% (High - 4/5 agree on country, 4/5 on city)
```

## Known Limitations

1. **Peer matching is imperfect** - Validator monikers don't always match node monikers
2. **Sentry nodes** - IPs may be sentries, not actual validator location
3. **VPNs/Proxies** - Some validators use VPNs which hide true location
4. **Cloud regions** - Cloud IPs may geolocate to company HQ, not actual region
5. **Rate limits** - Heavy usage may get rate-limited by free APIs
6. **Traceroute unreliability** - Most datacenter hosts block ICMP/UDP probes, so traceroute rarely reaches the destination and is discarded in those cases
7. **Anycast IPs** - CDN IPs (Cloudflare, etc.) geolocate to the nearest PoP, not the origin server

## Tips for Better Results

1. **Run during off-peak hours** to avoid API rate limits
2. **Check ASN info** - datacenter ASNs (Hetzner, AWS, GCP) indicate cloud hosting
3. **Website IPs are fallback** - they may point to CDN edges (Cloudflare, Vercel), not the validator itself

## CLI Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--percentage` | `-p` | `66.7` | Voting power percentage threshold |
| `--traceroute` | `-t` | `true` | Enable traceroute verification |
| `--json` | `-j` | `false` | Output results as JSON |

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `RPC_URL` | rpc-fuel-seq.simplystaking.xyz | Tendermint RPC endpoint |
| `REST_URL` | rest-fuel-seq.simplystaking.xyz | Cosmos REST API endpoint |

## Dependencies

**Go version:**
- Go 1.23+

**Shell version:**
- bash
- jq
- curl
- bc
- traceroute or mtr (for traceroute mode)
- whois (optional, for WHOIS lookups)

Install on Ubuntu/Debian:
```bash
apt-get install jq curl bc whois traceroute
```

Install on macOS:
```bash
brew install jq curl bc whois traceroute
```
