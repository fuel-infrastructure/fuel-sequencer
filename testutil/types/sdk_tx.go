package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/gogoproto/proto"
	"github.com/fuel-infrastructure/fuel-sequencer/utils"
)

func MustGetTxFromMsgs(msgs []proto.Message, sequence uint64) sdk.Tx {

	var msgAnys []*codectypes.Any
	for _, msg := range msgs {
		msgAny, err := codectypes.NewAnyWithValue(msg)
		if err != nil {
			panic(err)
		}
		msgAnys = append(msgAnys, msgAny)
	}

	bz, err := utils.ValidRawTxBytesFromAnyMsgs(msgAnys, sequence)
	if err != nil {
		panic(err)
	}

	sdkTx, err := tx.DefaultTxDecoder(TestCdc)(bz)
	if err != nil {
		panic(err)
	}

	return sdkTx
}
