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
	ErrInvalidEthAddressLength  = sdkerrors.Register(ModuleName, 1106,
		fmt.Sprintf("expected eth address to be %d bytes long", common.AddressLength),
	)
	ErrCouldNotDeserializeAuthorizeTx = sdkerrors.Register(ModuleName, 1107, "could not deserialize AuthorizeTx")
	ErrCouldExecuteMsg                = sdkerrors.Register(ModuleName, 1108, "could not execute msg")
)
