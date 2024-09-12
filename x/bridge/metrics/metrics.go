package metrics

import (
	"context"
	"strconv"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/utils"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	"github.com/hashicorp/go-metrics"
)

var (
	keysSequencer  = []string{"sequencer"}
	keysBeginBlock = append(keysSequencer, "begin", "block")
	keysTxMsg      = append(keysSequencer, "tx", "msg")
	keysStore      = append(keysSequencer, "store")
)

func ObserveDeposit(goCtx context.Context, depositedToUser bool) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
		telemetry.IncrCounterWithLabels(
			append(keysTxMsg, "deposit", "from", "ethereum"),
			1,
			[]metrics.Label{
				telemetry.NewLabel("to_user", strconv.FormatBool(depositedToUser)),
			},
		)
	})
}

func ObserveWithdrawal(goCtx context.Context) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
		telemetry.IncrCounter(1, append(keysTxMsg, "withdraw", "to", "ethereum")...)
	})
}

func ObserveSupplyDeltaForReport(goCtx context.Context, info types.SupplyDeltaInfo, currentSupply sdkmath.Int) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
		if info.LastSupply.IsInt64() && info.Offset.IsInt64() && currentSupply.IsInt64() {
			telemetry.SetGauge(float32(info.LastSupply.Int64()), append(keysBeginBlock, "supply", "last")...)
			telemetry.SetGauge(float32(info.Offset.Int64()), append(keysBeginBlock, "supply", "offset")...)
			telemetry.SetGauge(float32(currentSupply.Int64()), append(keysBeginBlock, "supply", "current")...)
		}
	})
}

func SetLastEthereumBlockSynced(goCtx context.Context, block uint64) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
		telemetry.SetGauge(float32(block), append(keysStore, "last", "ethereum", "block", "synced")...)
	})
}

func SetLastConsensusTxsSequence(goCtx context.Context, sequence uint64) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
		telemetry.SetGauge(float32(sequence), append(keysStore, "last", "consensus", "txs", "sequence")...)
	})
}

func SetLastEthBlockUpdate(goCtx context.Context, lastUpdate time.Time) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
		telemetry.SetGauge(float32(lastUpdate.Unix()), append(keysStore, "last", "eth", "block", "update")...)
	})
}

func SetLastEthereumNonce(goCtx context.Context, nonce sdkmath.Int) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
		if nonce.IsInt64() {
			telemetry.SetGauge(float32(nonce.Int64()), append(keysStore, "last", "ethereum", "nonce")...)
		}
	})
}

func SetEthereumEventIndexOffset(goCtx context.Context, offset uint64) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
		telemetry.SetGauge(float32(offset), append(keysStore, "ethereum", "event", "index", "offset")...)
	})
}

func SetNumInjectedTxsTotal(goCtx context.Context, txs uint64) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
		telemetry.SetGauge(float32(txs), append(keysStore, "index", "num", "injected", "txs", "total")...)
	})
}
