package basic_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (s *SpecialMsgsTestSuite) TestSpecialMsgsAuthorization() {

	s.Run("Ensure special messages cannot be submitted through an Authorize event", func() {

		// Set SupplyDeltaPeriod to a high number so that we can focus on our messages.
		bridgeParams := s.QueryBridgeParams(s.Ctx())
		bridgeParams.SupplyDeltaPeriod = 1000
		msgUpdateParams := bridgetypes.MsgUpdateParams{
			Authority: s.GetGovernanceAddress(),
			Params:    *bridgeParams,
		}
		s.ExecuteGovProposal(&msgUpdateParams)

		var msg sdk.Msg

		// Get starting block
		fromBlock, err := s.GetFuelSequencerHeight(s.Ctx())
		s.Require().NoError(err)

		// For a control test, let's try MsgUpdateParams first
		msg = &bridgetypes.MsgUpdateParams{
			Authority: testsuite.ETH_ADDRESSES[0],
			Params:    bridgetypes.DefaultParams(),
		}
		msgBz := s.GenerateMsgBz(msg)
		authorizeData := testsuite.PackAuthorize(msgBz)
		resp, err := s.SendEthTransactionToFuelStreamXContract(authorizeData)
		s.Require().NoError(err)

		// MsgIndex
		msg = &bridgetypes.MsgIndex{
			Authority:        testsuite.ADDRESSES[0],
			NumInjectedTxs:   0,
			NewEthereumBlock: false,
			BlockNumber:      s.QueryLastEthereumBlockSynced(s.Ctx()) + 1,
		}
		msgBz = s.GenerateMsgBz(msg)
		authorizeData = testsuite.PackAuthorize(msgBz)
		resp, err = s.SendEthTransactionToFuelStreamXContract(authorizeData)
		s.Require().NoError(err)

		// MsgDepositFromEthereum
		msg = &bridgetypes.MsgDepositFromEthereum{
			Authority: testsuite.ADDRESSES[0],
			Depositor: testsuite.ADDRESSES[0],
			Recipient: testsuite.ADDRESSES[1],
			Amount:    "1000",
			Lockup:    "0",
		}
		msgBz = s.GenerateMsgBz(msg)
		authorizeData = testsuite.PackAuthorize(msgBz)
		resp, err = s.SendEthTransactionToFuelStreamXContract(authorizeData)
		s.Require().NoError(err)

		// MsgSupplyDelta
		msg = &bridgetypes.MsgSupplyDelta{
			Authority: testsuite.ADDRESSES[0],
		}
		msgBz = s.GenerateMsgBz(msg)
		authorizeData = testsuite.PackAuthorize(msgBz)
		resp, err = s.SendEthTransactionToFuelStreamXContract(authorizeData)
		s.Require().NoError(err)

		// Wait for all messages to get processed
		s.PollForLastEthereumBlockSynced(s.Ctx(), 10, resp.BlockNumber.Uint64())

		// Get ending block
		toBlock, err := s.GetFuelSequencerHeight(s.Ctx())
		s.Require().NoError(err)

		// Check that no block has more than one transaction (the MsgIndex)
		for block := fromBlock; block <= toBlock; block++ {
			blockByHeight, err := s.GetBlockByHeight(s.Ctx(), int64(block))
			s.Require().NoError(err)
			s.Require().Len(blockByHeight.Data.Txs, 1)
		}
	})

	s.Run("Ensure special messages cannot be submitted by users", func() {
		var msg sdk.Msg

		// At CheckTx we do not apply any custom AnteHandler decorators. This means that the special messages will still
		// need to be valid by the standards of the default AnteHandler and also pass the message handler. However, we
		// then expect the messages to fail at block finalization where they will be subject to the custom decorator.

		// Pause Ethereum since we want to be able to have a more controlled environment for MsgIndex.
		s.PauseEthereum()

		// Set SupplyDeltaPeriod to 1 so the chain expects a SupplyDelta at every block and MsgSupplyDelta passes.
		bridgeParams := s.QueryBridgeParams(s.Ctx())
		bridgeParams.SupplyDeltaPeriod = 1
		msgUpdateParams := bridgetypes.MsgUpdateParams{
			Authority: s.GetGovernanceAddress(),
			Params:    *bridgeParams,
		}
		s.ExecuteGovProposal(&msgUpdateParams)

		// This is the error message we expect from every message.
		invalidAuthorityErrMsg := fmt.Sprintf("invalid authority; expected %s", s.GetGovernanceAddress())

		// For a control test, let's try MsgUpdateParams first
		msg = &bridgetypes.MsgUpdateParams{
			Authority: testsuite.ADDRESSES[0],
			Params:    bridgetypes.DefaultParams(),
		}
		resp, err := s.SubmitMsgs(msg)
		s.Require().NoError(err)
		s.Require().Contains(resp.RawLog, invalidAuthorityErrMsg)

		// MsgIndex
		msg = &bridgetypes.MsgIndex{
			Authority:        testsuite.ADDRESSES[0],
			NumInjectedTxs:   0,
			NewEthereumBlock: false,
			BlockNumber:      s.QueryLastEthereumBlockSynced(s.Ctx()) + 1,
		}
		resp, err = s.SubmitMsgs(msg)
		s.Require().NoError(err)
		s.Require().Contains(resp.RawLog, invalidAuthorityErrMsg)

		// MsgDepositFromEthereum
		msg = &bridgetypes.MsgDepositFromEthereum{
			Authority: testsuite.ADDRESSES[0],
			Depositor: testsuite.ADDRESSES[0],
			Recipient: testsuite.ADDRESSES[1],
			Amount:    "1000",
			Lockup:    "0",
		}
		resp, err = s.SubmitMsgs(msg)
		s.Require().NoError(err)
		s.Require().Contains(resp.RawLog, invalidAuthorityErrMsg)

		// MsgSupplyDelta
		msg = &bridgetypes.MsgSupplyDelta{
			Authority: testsuite.ADDRESSES[0],
		}
		resp, err = s.SubmitMsgs(msg)
		s.Require().NoError(err)
		s.Require().Contains(resp.RawLog, invalidAuthorityErrMsg)
	})
}
