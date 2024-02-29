package abci

/**
TODO: This file should be replaced as it was implemented for demonstration purposes
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
