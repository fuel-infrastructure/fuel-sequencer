package types

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
)

// x/bond module sentinel errors
var (
	ErrInvalidSigner = sdkerrors.Register(ModuleName, 1100, "invalid signer")
	ErrMintCoins     = sdkerrors.Register(ModuleName, 1101, "failed to mint coins")
	ErrSendCoins     = sdkerrors.Register(ModuleName, 1102, "failed to send coins")
)
