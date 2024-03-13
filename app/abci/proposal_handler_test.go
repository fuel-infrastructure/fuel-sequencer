package abci_test

func (s *AppTestSuite) TestPrepareProposalHandler() {
	testCases := []struct {
		name                       string
		setLastEthereumBlockSynced bool
		expErrMsg                  string
	}{
		{
			name:                       "error - LastEthereumBlockSynced not found",
			setLastEthereumBlockSynced: false,
			expErrMsg:                  "could not get last Ethereum block synced from state",
		},
	}

	//s.SetupTest()
	//ctrl := gomock.NewController(s.T())
	//defer ctrl.Finish()
	//sidecarClientMock := sidecartestutil.NewMockAppSidecarClient(ctrl)
	//sidecarClientMock.EXPECT().GetBlockEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("test error"))
	//propHandler := s.GetTestProposalHandler(sidecarClientMock)
	//propHandler.PrepareProposalHandler()(s.Ctx(), &abcitypes.RequestPrepareProposal{
	//	MaxTxBytes:         0,
	//	Txs:                nil,
	//	LocalLastCommit:    abcitypes.ExtendedCommitInfo{},
	//	Misbehavior:        nil,
	//	Height:             0,
	//	Time:               time.Time{},
	//	NextValidatorsHash: nil,
	//	ProposerAddress:    nil,
	//})

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			if tc.setLastEthereumBlockSynced {

			}
		})
	}
}
