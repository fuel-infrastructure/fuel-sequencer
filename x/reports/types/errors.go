package types

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
)

// Registered errors
var (
	ErrInvalidSigner        = sdkerrors.Register(ModuleName, 1100, "invalid signer")
	ErrSlashReportNotUnique = sdkerrors.Register(ModuleName, 1101, "duplicate slash report found")
	ErrSlashEntryNotUnique  = sdkerrors.Register(ModuleName, 1102, "duplicate slash entry found")
	ErrParamsInvalid        = sdkerrors.Register(ModuleName, 1103, "params are invalid")
)
