package metrics

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/telemetry"
	"github.com/hashicorp/go-metrics"
)

func ObserveDeposit(depositedToUser bool) {
	telemetry.IncrCounterWithLabels(
		[]string{"tx", "msg", "deposit_from_ethereum"},
		1,
		[]metrics.Label{
			telemetry.NewLabel("to_user", strconv.FormatBool(depositedToUser)),
		},
	)
}

func ObserveWithdrawal() {
	telemetry.IncrCounter(1, "tx", "msg", "withdraw_to_ethereum")
}
