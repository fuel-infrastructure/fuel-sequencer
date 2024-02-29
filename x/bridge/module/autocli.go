package bridge

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	modulev1 "github.com/fuel-infrastructure/fuel-sequencer/api/fuelsequencer/bridge"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: modulev1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Shows the parameters of the module",
				},
				{
					RpcMethod: "LastEthereumNonce",
					Use:       "show-last-ethereum-nonce",
					Short:     "show LastEthereumNonce",
				},
				{
					RpcMethod: "LastEthereumBlockSynced",
					Use:       "show-last-ethereum-block-synced",
					Short:     "show LastEthereumBlockSynced",
				},
				// this line is used by ignite scaffolding # autocli/query
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service:              modulev1.Msg_ServiceDesc.ServiceName,
			EnhanceCustomCommand: true, // only required if you want to use the custom command
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "UpdateParams",
					Skip:      true, // skipped because authority gated
				},
				{
					RpcMethod: "SupplyDelta",
					Skip:      true, // skipped because generated through concensus
				},
				{
					RpcMethod:      "WithdrawToEthereum",
					Use:            "withdraw-to-ethereum [to] [amount]",
					Short:          "Send a WithdrawToEthereum tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "to"}, {ProtoField: "amount"}},
				},
				// this line is used by ignite scaffolding # autocli/tx
			},
		},
	}
}
