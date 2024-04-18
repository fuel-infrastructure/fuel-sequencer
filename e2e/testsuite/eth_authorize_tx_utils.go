package testsuite

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/gogoproto/proto"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (s *E2ETestSuite) GenerateMsgSendBytes(fromAddress, toAddress string, amount []sdk.Coin) []byte {
	msgSend := &banktypes.MsgSend{
		FromAddress: fromAddress,
		ToAddress:   toAddress,
		Amount:      amount,
	}
	anyMsgSend, err := codectypes.NewAnyWithValue(msgSend)
	s.Require().NoError(err)

	// Serialize the message to bytes
	bz, err := proto.Marshal(&bridgetypes.AuthorizeTx{Messages: []*codectypes.Any{anyMsgSend}})
	s.Require().NoError(err)
	return bz
}
