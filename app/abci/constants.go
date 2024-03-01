package abci

const (
	// MaxVESize defines the maximum size of a vote extension in bytes.
	// TODO: We should use concise codecs and afterwards benchmark the application to set this value accordingly.
	//     : QA benchmarks here: https://github.com/cometbft/cometbft/blob/v0.38.0-rc1/docs/qa/CometBFT-QA-38.md#vote-extensions-testbed
	MaxVESize = 1024
)
