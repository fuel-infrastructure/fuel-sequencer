package types

import (
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	// this line is used by starport scaffolding # 1
)

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {

	// This is telling Cosmos SDK that it should recognise these accounts as implementations of the account interfaces.
	// Otherwise, if it's unmarshalling (e.g. while loading the genesis file) it will not recognise the account types.
	registry.RegisterImplementations((*sdk.AccountI)(nil),
		&EthOwnedBaseAccount{},
		&EthOwnedContinuousVestingAccount{},
		&EthOwnedMultiContinuousVestingAccount{},
	)
	registry.RegisterImplementations((*authtypes.GenesisAccount)(nil),
		&EthOwnedBaseAccount{},
		&EthOwnedContinuousVestingAccount{},
		&EthOwnedMultiContinuousVestingAccount{},
	)

	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUpdateParams{},
		&MsgSupplyDelta{},
		&MsgWithdrawToEthereum{},
		&MsgDepositFromEthereum{},
		&MsgIndex{},
	)
	// this line is used by starport scaffolding # 3

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
