package metrics

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/telemetry"
	"github.com/hashicorp/go-metrics"
)

var (
	keysProcessProposal = []string{"abci", "process_proposal"}
)

func ProcessProposalResponse(accept bool) {
	telemetry.IncrCounterWithLabels(
		append(keysProcessProposal, "response"),
		1,
		[]metrics.Label{telemetry.NewLabel("accept", strconv.FormatBool(accept))},
	)
}

func SkipAuthorizeEvent() {
	telemetry.IncrCounter(1, append(keysProcessProposal, "skip", "authorize", "event")...)
}
