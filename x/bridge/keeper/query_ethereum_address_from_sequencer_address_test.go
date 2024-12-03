package keeper_test

import (
	testutiltypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (s *KeeperTestSuite) TestEthereumAddressFromSequencerAddress() {

	testCases := []struct {
		name                  string
		sequencerAddress      string
		expectEthereumAddress string
		expectErrMsg          string
	}{
		{
			name:                  "valid sequencer account address => expected ethereum address",
			sequencerAddress:      testutiltypes.TestSeqAddr1Str,
			expectEthereumAddress: testutiltypes.TestEthAddr1Str,
		},
		{
			name:                  "valid sequencer validator address => expected ethereum address",
			sequencerAddress:      testutiltypes.TestValAddr1Str,
			expectEthereumAddress: testutiltypes.TestEthAddr1Str,
		},
		{
			name:             "invalid sequencer address => err",
			sequencerAddress: "invalid_sequencer_address",
			expectErrMsg:     "could not parse address into a Sequencer address: decoding bech32 failed",
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			req := &types.QueryEthereumAddressFromSequencerAddressRequest{SequencerAddress: tc.sequencerAddress}
			resp, err := s.App.BridgeKeeper.EthereumAddressFromSequencerAddress(s.Ctx(), req)
			if tc.expectErrMsg != "" {
				s.Require().ErrorContains(err, tc.expectErrMsg)
				return
			}
			s.Require().NoError(err)

			s.Require().Equal(tc.expectEthereumAddress, resp.EthereumAddress)
		})
	}
}
