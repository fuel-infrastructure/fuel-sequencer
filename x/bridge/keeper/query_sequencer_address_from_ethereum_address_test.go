package keeper_test

import (
	testutiltypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (s *KeeperTestSuite) TestSequencerAddressFromEthereumAddress() {

	testCases := []struct {
		name                   string
		ethereumAddress        string
		expectSequencerAddress string
		expectErrMsg           string
	}{
		{
			name:                   "valid ethereum address => expected sequencer address",
			ethereumAddress:        testutiltypes.TestEthAddr1Str,
			expectSequencerAddress: testutiltypes.TestSeqAddr1Str,
		},
		{
			name:            "invalid ethereum address => err",
			ethereumAddress: "invalid_ethereum_address",
			expectErrMsg:    "invalid Ethereum address format (invalid_ethereum_address)",
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			req := &types.QuerySequencerAddressFromEthereumAddressRequest{EthereumAddress: tc.ethereumAddress}
			resp, err := s.App.BridgeKeeper.SequencerAddressFromEthereumAddress(s.Ctx(), req)
			if tc.expectErrMsg != "" {
				s.Require().ErrorContains(err, tc.expectErrMsg)
				return
			}
			s.Require().NoError(err)

			s.Require().Equal(tc.expectSequencerAddress, resp.SequencerAddress)
		})
	}
}
