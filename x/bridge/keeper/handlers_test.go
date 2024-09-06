package keeper_test

// This test confirms that the Bridge module's messages were registered with the correct Type URL
func (s *KeeperTestSuite) TestMessagesRegisteredWithCorrectTypeUrl() {
	handler := s.App.MsgServiceRouter().HandlerByTypeURL("/fuelsequencer.bridge.v1.MsgUpdateParams")
	s.Require().NotNil(handler)
	handler = s.App.MsgServiceRouter().HandlerByTypeURL("/fuelsequencer.bridge.v1.MsgSupplyDelta")
	s.Require().NotNil(handler)
	handler = s.App.MsgServiceRouter().HandlerByTypeURL("/fuelsequencer.bridge.v1.MsgWithdrawToEthereum")
	s.Require().NotNil(handler)
	handler = s.App.MsgServiceRouter().HandlerByTypeURL("/fuelsequencer.bridge.v1.MsgDepositFromEthereum")
	s.Require().NotNil(handler)
	handler = s.App.MsgServiceRouter().HandlerByTypeURL("/fuelsequencer.bridge.v1.MsgIndex")
	s.Require().NotNil(handler)
}

// This test confirms that the Bridge module's queries were registered with the correct path
func (s *KeeperTestSuite) TestQueriesRegisteredWithCorrectPath() {
	handler := s.App.GRPCQueryRouter().Route("/fuelsequencer.bridge.v1.Query/Params")
	s.Require().NotNil(handler)
	handler = s.App.GRPCQueryRouter().Route("/fuelsequencer.bridge.v1.Query/LastEthereumNonce")
	s.Require().NotNil(handler)
	handler = s.App.GRPCQueryRouter().Route("/fuelsequencer.bridge.v1.Query/LastEthereumBlockSynced")
	s.Require().NotNil(handler)
	handler = s.App.GRPCQueryRouter().Route("/fuelsequencer.bridge.v1.Query/EthereumEventIndexOffset")
	s.Require().NotNil(handler)
	handler = s.App.GRPCQueryRouter().Route("/fuelsequencer.bridge.v1.Query/SupplyDeltaInfo")
	s.Require().NotNil(handler)
	handler = s.App.GRPCQueryRouter().Route("/fuelsequencer.bridge.v1.Query/SequencerAddressFromEthereumAddress")
	s.Require().NotNil(handler)
	handler = s.App.GRPCQueryRouter().Route("/fuelsequencer.bridge.v1.Query/LastEthBlockUpdateTime")
	s.Require().NotNil(handler)
	handler = s.App.GRPCQueryRouter().Route("/fuelsequencer.bridge.v1.Query/LastConsensusTxsSequence")
	s.Require().NotNil(handler)
}
