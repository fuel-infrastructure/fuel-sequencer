package types

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
)

// x/bridge module sentinel errors
var (
	ErrInvalidSigner = sdkerrors.Register(
		ModuleName, 1100, "expected gov account as only signer for proposal message",
	)
	ErrInvalidSupplyDeltaPeriod = sdkerrors.Register(ModuleName, 1101, "invalid param SupplyDeltaPeriod")
	ErrUnexpectedOperation      = sdkerrors.Register(ModuleName, 1102, "operation was not expected")
	ErrInvalidSupplyDeltaValue  = sdkerrors.Register(ModuleName, 1103, "supply delta value is invalid")
	ErrInvalidEthAddress        = sdkerrors.Register(ModuleName, 1104, "invalid ethereum address")
)
