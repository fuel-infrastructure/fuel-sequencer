package keeper_test

// This test confirms that the Sequencing module's messages were registered with the correct Type URL
func (s *KeeperTestSuite) TestMessagesRegisteredWithCorrectTypeUrl() {
	handler := s.App.MsgServiceRouter().HandlerByTypeURL("/fuelsequencer.sequencing.v1.MsgUpdateParams")
	s.Require().NotNil(handler)
	handler = s.App.MsgServiceRouter().HandlerByTypeURL("/fuelsequencer.sequencing.v1.MsgPostBlob")
	s.Require().NotNil(handler)
}

// This test confirms that the Sequencing module's queries were registered with the correct path
func (s *KeeperTestSuite) TestQueriesRegisteredWithCorrectPath() {
	handler := s.App.GRPCQueryRouter().Route("/fuelsequencer.sequencing.v1.Query/Params")
	s.Require().NotNil(handler)
	handler = s.App.GRPCQueryRouter().Route("/fuelsequencer.sequencing.v1.Query/Topic")
	s.Require().NotNil(handler)
	handler = s.App.GRPCQueryRouter().Route("/fuelsequencer.sequencing.v1.Query/TopicAll")
	s.Require().NotNil(handler)
}
