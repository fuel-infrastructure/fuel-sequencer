package types

// DONTCOVER

import (
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	"github.com/ethereum/go-ethereum/common"
)

var (
	ErrInvalidSigner = sdkerrors.Register(
		ModuleName, 1100, "expected gov account as only signer for proposal message",
	)
	ErrInvalidSupplyDeltaPeriod = sdkerrors.Register(ModuleName, 1101, "invalid param SupplyDeltaPeriod")
	ErrUnexpectedOperation      = sdkerrors.Register(ModuleName, 1102, "operation was not expected")
	ErrInvalidSupplyDeltaValue  = sdkerrors.Register(ModuleName, 1103, "supply delta value is invalid")
	ErrInvalidEthAddress        = sdkerrors.Register(ModuleName, 1104, "invalid ethereum address")
	ErrInvalidVestingDuration   = sdkerrors.Register(ModuleName, 1105, "invalid vesting duration")
	ErrInvalidEthAddressLength  = sdkerrors.Register(
		ModuleName, 1106, fmt.Sprintf("expected eth address to be %d bytes long", common.AddressLength),
	)
	ErrEthEventsTxBlockNotSequential = sdkerrors.Register(
		ModuleName, 1107, "eth events tx block number is not the increment of last ethereum block synced")
	ErrUnsupported           = sdkerrors.Register(ModuleName, 1108, "unsupported")
	ErrInvalidAccountAddress = sdkerrors.Register(ModuleName, 1109, "invalid account address")
	ErrAccountOwnerMismatch  = sdkerrors.Register(ModuleName, 1110, "account owner mismatch")
	ErrParamsInvalid         = sdkerrors.Register(ModuleName, 1111, "params are invalid")
)
