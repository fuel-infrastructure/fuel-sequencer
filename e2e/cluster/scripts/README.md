# Genesis Account Mnemonic Generator

This directory contains utilities for generating valid BIP-39 mnemonics for additional genesis accounts.

## Generating Mnemonics

To generate additional valid BIP-39 mnemonics for genesis accounts:

```bash
# Generate 60 mnemonics (12 words each)
go run generate_mnemonics.go -count 60

# Generate 10 mnemonics with 24 words each
go run generate_mnemonics.go -count 10 -bits 256
```

### Parameters

- `-count`: Number of mnemonics to generate (default: 60)
- `-bits`: Entropy bits - 128 for 12-word mnemonics, 256 for 24-word mnemonics (default: 128)

## Usage

1. Run the generator to create mnemonics
2. Copy the output mnemonics
3. Add them to the `AdditionalGenesisMnemonics` slice in `e2e/cluster/internal/sequencer/parameters.go`
4. The accounts will be automatically created at genesis with the configured balance

## Example

```bash
$ go run generate_mnemonics.go -count 3
// Generated valid BIP-39 mnemonics
		"shock rail nothing gloom run kangaroo forget affair they remove company fatal",
		"misery muscle brown render drink evolve latin spy mention guitar spell fever",
		"ancient model beef staff hole volcano job gas scan skirt kid tooth",
```

## Configuration

The balance for additional genesis accounts is configured in `parameters.go`:

```go
additionalAccountBalance = sdkmath.NewIntFromString("1000000000000000000") // 1 bil x 1e9
```

You can modify this value to change the initial balance for all additional accounts.
