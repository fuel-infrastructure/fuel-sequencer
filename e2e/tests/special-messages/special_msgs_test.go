package basic_test

import (
	"fmt"

	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (s *SpecialMsgsTestSuite) TestSpecialMsgsAuthorization() {

	s.Run("Ensure special messages cannot be submitted by users", func() {

		// At CheckTx we do not apply any custom AnteHandler decorators. This means that the special messages will still
		// need to be valid by the standards of the default AnteHandler and also pass the message handler. However, we
		// then expect the messages to fail at block finalization where they will be subject to the custom decorator.

		// Pause Ethereum since we want to be able to have a more controlled environment for MsgIndex.
		s.PauseEthereum()

		// Set SupplyDeltaPeriod to 1 so the chain expects a SupplyDelta at every block and MsgSupplyDelta passes.
		bridgeParams := s.QueryBridgeParams(s.Ctx())
		bridgeParams.SupplyDeltaPeriod = 1
		msgUpdateParams := &bridgetypes.MsgUpdateParams{
			Authority: s.GetGovernanceAddress(),
			Params:    *bridgeParams,
		}
		s.ExecuteGovProposal(msgUpdateParams)

		// This is the error message we expect from every message.
		invalidAuthorityErrMsg := fmt.Sprintf("invalid authority; expected %s", s.GetGovernanceAddress())

		// ------------ MsgIndex

		msgIndex := &bridgetypes.MsgIndex{
			Authority:           s.SeqKeys[0].AddressSeq,
			NumInjectedEventTxs: 0,
			NewEthereumBlock:    false,
			BlockNumber:         s.QueryLastEthereumBlockSynced(s.Ctx()) + 1,
		}
		resp, err := s.SubmitMsgs(msgIndex)
		s.Require().NoError(err)
		s.Require().Contains(resp.RawLog, invalidAuthorityErrMsg)

		// ------------ MsgDepositFromEthereum

		msgDepositFromEthereum := &bridgetypes.MsgDepositFromEthereum{
			Authority: s.SeqKeys[0].AddressSeq,
			Depositor: s.SeqKeys[0].AddressSeq,
			Recipient: s.SeqKeys[1].AddressSeq,
			Amount:    "1000",
			Lockup:    "0",
		}
		resp, err = s.SubmitMsgs(msgDepositFromEthereum)
		s.Require().NoError(err)
		s.Require().Contains(resp.RawLog, invalidAuthorityErrMsg)

		// ------------ MsgSupplyDelta

		msgSupplyDelta := &bridgetypes.MsgSupplyDelta{
			Authority: s.SeqKeys[0].AddressSeq,
		}
		resp, err = s.SubmitMsgs(msgSupplyDelta)
		s.Require().NoError(err)
		s.Require().Contains(resp.RawLog, invalidAuthorityErrMsg)
	})
}
