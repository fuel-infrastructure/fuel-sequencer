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

func ObserveDeposit(goCtx context.Context, depositedToUser bool) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
		telemetry.IncrCounterWithLabels(
			append(utils.KeysTxMsg, "deposit", "from", "ethereum"),
			1,
			[]metrics.Label{
				telemetry.NewLabel("to_user", strconv.FormatBool(depositedToUser)),
			},
		)
	})
}

func ObserveWithdrawal(goCtx context.Context) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
		telemetry.IncrCounter(1, append(utils.KeysTxMsg, "withdraw", "to", "ethereum")...)
	})
}

func ObserveSupplyDeltaForReport(goCtx context.Context, info types.SupplyDeltaInfo, currentSupply sdkmath.Int) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
		telemetry.SetGauge(utils.ScaleCoinAmount(info.LastSupply), append(utils.KeysBeginBlock, "supply", "last")...)
		telemetry.SetGauge(utils.ScaleCoinAmount(info.Offset), append(utils.KeysBeginBlock, "supply", "offset")...)
		telemetry.SetGauge(utils.ScaleCoinAmount(currentSupply), append(utils.KeysBeginBlock, "supply", "current")...)
	})
}

func SetLastEthereumBlockSynced(goCtx context.Context, block uint64) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
		telemetry.SetGauge(float32(block), append(utils.KeysStore, "last", "ethereum", "block", "synced")...)
	})
}

func SetLastConsensusTxsSequence(goCtx context.Context, sequence uint64) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
		telemetry.SetGauge(float32(sequence), append(utils.KeysStore, "last", "consensus", "txs", "sequence")...)
	})
}

func SetLastEthBlockUpdate(goCtx context.Context, lastUpdate time.Time) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
		telemetry.SetGauge(float32(lastUpdate.Unix()), append(utils.KeysStore, "last", "eth", "block", "update")...)
	})
}

func SetLastEthereumNonce(goCtx context.Context, nonce sdkmath.Int) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
		if nonce.IsInt64() {
			telemetry.SetGauge(float32(nonce.Int64()), append(utils.KeysStore, "last", "ethereum", "nonce")...)
		}
	})
}

func SetEthereumEventIndexOffset(goCtx context.Context, offset uint64) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
		telemetry.SetGauge(float32(offset), append(utils.KeysStore, "ethereum", "event", "index", "offset")...)
	})
}

func SetNumInjectedTxsTotal(goCtx context.Context, txs uint64) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
		telemetry.SetGauge(float32(txs), append(utils.KeysStore, "index", "num", "injected", "txs", "total")...)
	})
}
