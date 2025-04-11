package types

import (
	"context"

	"cosmossdk.io/math"
)

// ParamSubspace defines the expected Subspace interface for parameters.
type ParamSubspace interface {
	Get(ctx context.Context, key []byte, ptr interface{})
	Set(ctx context.Context, key []byte, param interface{})
	Has(ctx context.Context, key []byte) bool
}

// BridgeKeeper defines the expected interface for the Bridge module
type BridgeKeeper interface {
	SetLastEthereumNonce(ctx context.Context, nonce math.Int)
	MustGetLastEthereumNonce(ctx context.Context) math.Int
}
