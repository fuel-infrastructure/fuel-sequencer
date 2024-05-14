package bridge

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	modulev1 "github.com/fuel-infrastructure/fuel-sequencer/api/fuelsequencer/bridge/v1"
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
					RpcMethod: "SupplyDeltaInfo",
					Use:       "show-supply-delta-info",
					Short:     "show supply-delta-info",
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
				{
					RpcMethod:      "SequencerAddressFromEthereumAddress",
					Use:            "sequencer-address-from-ethereum-address [ethereum-address]",
					Short:          "Query SequencerAddressFromEthereumAddress",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "ethereum_address"}},
				},
				{
					RpcMethod: "SupplyDeltaProcessed",
					Use:       "show-supply-delta-processed",
					Short:     "show supply-delta-processed",
				},
				{
					RpcMethod: "EthereumEventIndexOffset",
					Use:       "show-ethereum-event-index-offset",
					Short:     "show ethereum-event-index-offset",
				},
				{
					RpcMethod: "LastEthBlockUpdateTime",
					Use:       "show-last-eth-block-update-time",
					Short:     "show last-eth-block-update-time",
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
					Skip:      true, // skipped because generated through consensus
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
