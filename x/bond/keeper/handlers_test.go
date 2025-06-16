package keeper_test

// This test confirms that the Bond module's messages were registered with the correct Type URL
func (s *KeeperTestSuite) TestMessagesRegisteredWithCorrectTypeUrl() {
	handler := s.App.MsgServiceRouter().HandlerByTypeURL("/fuelsequencer.bond.v1.MsgUpdateParams")
	s.Require().NotNil(handler)
}

// This test confirms that the Bond module's queries were registered with the correct path
func (s *KeeperTestSuite) TestQueriesRegisteredWithCorrectPath() {
	handler := s.App.GRPCQueryRouter().Route("/fuelsequencer.bond.v1.Query/Params")
	s.Require().NotNil(handler)
}
