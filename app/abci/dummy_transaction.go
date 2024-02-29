package abci

/**
TODO: This file should be replaced as it was implemented for demonstration purposes. Please note that here we are using
    : JSON to marshal the vote extension for simplicity. However,the Cosmos SDK docs suggest using a more lightweight
    : encoding that produce a small output, such as compressed bytes or custom encodings.
*/

import (
	"encoding/json"

	abci "github.com/cometbft/cometbft/abci/types"
)

type AggregatedOracleData struct {
	ParsedOracleData   string
	Msgs               string
	ExtendedCommitInfo abci.ExtendedCommitInfo
}

func (p *AggregatedOracleData) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

func (p *AggregatedOracleData) Unmarshal(bz []byte) error {
	return json.Unmarshal(bz, p)
}
