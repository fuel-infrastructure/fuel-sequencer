package testsuite

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gogoproto/proto"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (s *E2ETestSuite) generateMsgBytes(msg sdk.Msg) []byte {
	anyMsgSend, err := codectypes.NewAnyWithValue(msg)
	s.Require().NoError(err)

	// Serialize the message to bytes
	bz, err := proto.Marshal(&bridgetypes.AuthorizeTx{Messages: []*codectypes.Any{anyMsgSend}})
	s.Require().NoError(err)
	return bz
}

func (s *E2ETestSuite) GenerateMsgSendBytes(fromAddress, toAddress string, amount []sdk.Coin) []byte {
	return s.generateMsgBytes(
		&banktypes.MsgSend{
			FromAddress: fromAddress,
			ToAddress:   toAddress,
			Amount:      amount,
		},
	)
}

func (s *E2ETestSuite) GenerateMsgDelegateBytes(delegatorAddress, validatorAddress string, amount sdk.Coin) []byte {
	return s.generateMsgBytes(
		&stakingtypes.MsgDelegate{
			DelegatorAddress: delegatorAddress,
			ValidatorAddress: validatorAddress,
			Amount:           amount,
		},
	)
}

func (s *E2ETestSuite) GenerateMsgBeginRedelegateBytes(
	delegatorAddress, validatorSrcAddress, validatorDstAddress string, amount sdk.Coin,
) []byte {
	return s.generateMsgBytes(
		&stakingtypes.MsgBeginRedelegate{
			DelegatorAddress:    delegatorAddress,
			ValidatorSrcAddress: validatorSrcAddress,
			ValidatorDstAddress: validatorDstAddress,
			Amount:              amount,
		},
	)
}

func (s *E2ETestSuite) GenerateMsgWithdrawDelegatorRewardBytes(delegatorAddress, validatorAddress string) []byte {
	return s.generateMsgBytes(
		&distributiontypes.MsgWithdrawDelegatorReward{
			DelegatorAddress: delegatorAddress,
			ValidatorAddress: validatorAddress,
		},
	)
}

func (s *E2ETestSuite) GenerateMsgUndelegateBytes(delegatorAddress, validatorAddress string, amount sdk.Coin) []byte {
	return s.generateMsgBytes(
		&stakingtypes.MsgUndelegate{
			DelegatorAddress: delegatorAddress,
			ValidatorAddress: validatorAddress,
			Amount:           amount,
		},
	)
}
