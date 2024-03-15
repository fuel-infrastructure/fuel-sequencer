package basic_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
)

func (s *BasicTestSuite) TestStartUpAndBasicQueries() {
	s.Run("Bring up nodes and perform some queries and transactions", func() {

		// Try getting module params (GRPC).
		params := s.QueryBankParams(s.Ctx())
		s.Require().True(params.DefaultSendEnabled)

		// Try getting address balances (GRPC).
		balance, err := s.QueryAllBalances(s.Ctx(), testsuite.ADDRESSES[0], nil)
		s.Require().NoError(err)
		expected := testsuite.InitBalanceCoin.Sub(testsuite.InitStakedCoin)
		s.Require().True(expected.Amount.Equal(balance.Balances.AmountOf(testsuite.BridgeDenom)))

		// Try getting height (RPC).
		height, err := s.GetFuelSequencerHeight(s.Ctx())
		s.Require().NoError(err)
		s.Require().Greater(height, uint64(1))

		// Try querying block by height (GRPC).
		block, err := s.GetBlockByHeight(s.Ctx(), height)
		s.Require().NoError(err)
		s.Require().Equal(height, uint64(block.Header.Height))

		// Try transferring tokens (RPC).
		from := sdk.MustAccAddressFromBech32(testsuite.ADDRESSES[0])
		to := sdk.MustAccAddressFromBech32(testsuite.ADDRESSES[1])
		amount := sdk.NewCoins(sdk.NewInt64Coin(testsuite.BridgeDenom, 100))
		msg := banktypes.NewMsgSend(from, to, amount)
		res, err := s.SubmitMsgs(msg)
		s.Require().NoError(err)
		s.Require().Zero(res.Code)

		// Wait for blocks (RPC).
		err = s.WaitForBlocks(s.Ctx(), 2, time.Minute)
		s.Require().NoError(err)

		// Ensure balance was reduced (GRPC)
		// Note: a fee was also charged.
		updatedBalance, err := s.QueryAllBalances(s.Ctx(), testsuite.ADDRESSES[0], nil)
		s.Require().NoError(err)
		s.Require().True(updatedBalance.Balances.IsAllLT(balance.Balances))

		// Ensure we can query the sidecar (GRPC)
		//
		// ...we do not expect events at height 1.
		events, err := s.QuerySidecarBlockEvents(s.Ctx(), 1)
		s.Require().NoError(err)
		s.Require().Empty(events)
		//
		// ...we expect events at height 4.
		events, err = s.QuerySidecarBlockEvents(s.Ctx(), 4)
		s.Require().NoError(err)
		s.Require().NotEmpty(events)
		//
		// ...we expect an error if we query a block that doesn't exist.
		events, err = s.QuerySidecarBlockEvents(s.Ctx(), 10)
		s.Require().Error(err)

		// TODO: Ensure that the Sequencer is synced up. (once we have PreBlocker logic)
		//err = s.WaitForBlocks(s.Ctx(), 5, time.Minute)
		//s.Require().NoError(err)
		//lastEthereumBlockSynced := s.QueryLastEthereumBlockSynced(s.Ctx())
		//s.Require().Equal(4, lastEthereumBlockSynced)
	})
}
