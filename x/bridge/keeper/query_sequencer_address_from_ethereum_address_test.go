package keeper_test

import (
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
			ethereumAddress:        "0x71C7656EC7ab88b098defB751B7401B5f6d8976F",
			expectSequencerAddress: "fuelsequencer13tch2uhman7dhjjphmx9uwx7kvg2kqfj5y56hsmljlv93pgma5vqyks99k",
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
