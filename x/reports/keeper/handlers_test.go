package keeper_test

// This test confirms that the Reports module's messages were registered with the correct Type URL
func (s *KeeperTestSuite) TestMessagesRegisteredWithCorrectTypeUrl() {
	handler := s.App.MsgServiceRouter().HandlerByTypeURL("/fuelsequencer.reports.v1.MsgUpdateParams")
	s.Require().NotNil(handler)
}

// This test confirms that the Reports module's queries were registered with the correct path
func (s *KeeperTestSuite) TestQueriesRegisteredWithCorrectPath() {
	handler := s.App.GRPCQueryRouter().Route("/fuelsequencer.reports.v1.Query/Params")
	s.Require().NotNil(handler)
	handler = s.App.GRPCQueryRouter().Route("/fuelsequencer.reports.v1.Query/SlashReport")
	s.Require().NotNil(handler)
	handler = s.App.GRPCQueryRouter().Route("/fuelsequencer.reports.v1.Query/SlashReportAll")
	s.Require().NotNil(handler)
}
