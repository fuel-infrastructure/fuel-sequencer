package testsuite

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gogoproto/proto"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (s *E2ETestSuite) GenerateMsgBz(msg sdk.Msg) []byte {
	anyMsgSend, err := codectypes.NewAnyWithValue(msg)
	s.Require().NoError(err)

	// Serialize the message to bytes
	bz, err := proto.Marshal(&bridgetypes.AuthorizeTx{Messages: []*codectypes.Any{anyMsgSend}})
	s.Require().NoError(err)
	return bz
}

func (s *E2ETestSuite) GenerateMsgSendBz(fromAddress, toAddress string, amount []sdk.Coin) []byte {
	return s.GenerateMsgBz(
		&banktypes.MsgSend{
			FromAddress: fromAddress,
			ToAddress:   toAddress,
			Amount:      amount,
		},
	)
}

func (s *E2ETestSuite) GenerateMsgDelegateBz(delegatorAddress, validatorAddress string, amount sdk.Coin) []byte {
	return s.GenerateMsgBz(
		&stakingtypes.MsgDelegate{
			DelegatorAddress: delegatorAddress,
			ValidatorAddress: validatorAddress,
			Amount:           amount,
		},
	)
}

func (s *E2ETestSuite) GenerateMsgBeginRedelegateBz(
	delegatorAddress, validatorSrcAddress, validatorDstAddress string, amount sdk.Coin,
) []byte {
	return s.GenerateMsgBz(
		&stakingtypes.MsgBeginRedelegate{
			DelegatorAddress:    delegatorAddress,
			ValidatorSrcAddress: validatorSrcAddress,
			ValidatorDstAddress: validatorDstAddress,
			Amount:              amount,
		},
	)
}

func (s *E2ETestSuite) GenerateMsgWithdrawDelegatorRewardBz(delegatorAddress, validatorAddress string) []byte {
	return s.GenerateMsgBz(
		&distributiontypes.MsgWithdrawDelegatorReward{
			DelegatorAddress: delegatorAddress,
			ValidatorAddress: validatorAddress,
		},
	)
}

func (s *E2ETestSuite) GenerateMsgUndelegateBz(delegatorAddress, validatorAddress string, amount sdk.Coin) []byte {
	return s.GenerateMsgBz(
		&stakingtypes.MsgUndelegate{
			DelegatorAddress: delegatorAddress,
			ValidatorAddress: validatorAddress,
			Amount:           amount,
		},
	)
}

func (s *E2ETestSuite) GenerateMsgWithdrawToEthereumBz(from, to string, amount sdk.Coin) []byte {
	return s.GenerateMsgBz(
		&bridgetypes.MsgWithdrawToEthereum{
			From:   from,
			To:     to,
			Amount: amount,
		},
	)
}

func (s *E2ETestSuite) GenerateMsgVoteBz(
	proposalId uint64, voter, metadata string, option govtypesv1.VoteOption,
) []byte {
	return s.GenerateMsgBz(
		&govtypesv1.MsgVote{
			ProposalId: proposalId,
			Voter:      voter,
			Option:     option,
			Metadata:   metadata,
		},
	)
}
