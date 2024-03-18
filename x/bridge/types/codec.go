package types

import (
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	// this line is used by starport scaffolding # 1
)

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.AccountI)(nil),
		&EthOwnedBaseAccount{},
		&EthOwnedContinuousVestingAccount{},
	)
	registry.RegisterImplementations((*authtypes.GenesisAccount)(nil),
		&EthOwnedBaseAccount{},
		&EthOwnedContinuousVestingAccount{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUpdateParams{},
		&MsgSupplyDelta{},
		&MsgWithdrawToEthereum{},
	)
	// this line is used by starport scaffolding # 3

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
