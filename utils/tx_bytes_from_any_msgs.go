package utils

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/gogoproto/proto"
)

const (
	// InjectedTxGasLimit is the default gas limit set to injected transactions. This value is set to zero because
	// inside the ante handler we are overriding with an infinite gas meter to make sure that the messages do not
	// fail due to insufficient gas.
	InjectedTxGasLimit = uint64(0)
)

// ValidRawTxBytesFromAnyMsgs encodes a transaction that will be injected by consensus and considered as valid by the
// Cosmos SDK transaction decoder. This is important for the transaction to produce a transaction result, meaning that
// it will contribute to the block's last results hash and thus be provable on Ethereum via Bridge Commitments.
//
// The transaction's gas limit is set to zero with the assumption that an infinite gas meter will be used.
func ValidRawTxBytesFromAnyMsgs(msgs []*codectypes.Any, sequence uint64) ([]byte, error) {

	// Construct Tx Body with the message.
	txBodyBz, err := proto.Marshal(&txtypes.TxBody{
		Messages: msgs,
	})
	if err != nil {
		return nil, err
	}

	// Construct Auth Info with Fee to avoid nil pointer panics.
	authInfoBz, err := proto.Marshal(&txtypes.AuthInfo{
		SignerInfos: []*txtypes.SignerInfo{
			{Sequence: sequence},
		},
		Fee: &txtypes.Fee{
			GasLimit: InjectedTxGasLimit,
		},
	})
	if err != nil {
		return nil, err
	}

	// Construct final Tx.
	txRawBz, err := proto.Marshal(&txtypes.TxRaw{
		BodyBytes:     txBodyBz,
		AuthInfoBytes: authInfoBz,
		Signatures:    nil,
	})
	if err != nil {
		return nil, err
	}

	return txRawBz, nil
}
