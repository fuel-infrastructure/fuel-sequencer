package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	testtypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
)

func (s *KeeperTestSuite) TestBurnCoinsFromAddress() {

	withdrawer := s.TestAccs[0]
	denom := "fuel"
	amount := sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(200)))

	testCases := []struct {
		name         string
		fundAccounts bool
		expErrMsg    string
	}{
		{
			"successfully burn coins from address",
			true,
			"",
		},
		{
			"error sending coins from account to module",
			false,
			fmt.Sprintf("cannot send tokens from %s to bridge module: spendable balance 0fuel is smaller than 200fuel", withdrawer),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			if tc.fundAccounts {
				s.FundAcc(s.Ctx(), withdrawer, amount)
			}

			totalSupplyBefore := s.App.BankKeeper.GetSupply(s.Ctx(), denom)
			withdrawBalanceBefore := s.App.BankKeeper.GetAllBalances(s.Ctx(), withdrawer)

			err := s.App.BridgeKeeper.BurnCoinsFromAddress(s.Ctx(), withdrawer, amount)

			totalSupplyAfter := s.App.BankKeeper.GetSupply(s.Ctx(), denom)
			withdrawBalanceAfter := s.App.BankKeeper.GetAllBalances(s.Ctx(), withdrawer)

			if len(tc.expErrMsg) > 0 {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
				s.Require().Equal(totalSupplyBefore, totalSupplyAfter)
				s.Require().Equal(withdrawBalanceBefore, withdrawBalanceAfter)
				return
			}

			s.Require().NoError(err)
			s.Require().True(totalSupplyAfter.IsZero())
			s.Require().True(withdrawBalanceAfter.IsZero())
		})
	}
}

func (s *KeeperTestSuite) TestDeserializeAuthorizeTx() {
	amt, ok := sdkmath.NewIntFromString("10")
	s.Require().True(ok)

	testCases := []struct {
		name      string
		event     *sidecartypes.AuthorizeEvent
		expMsgs   []sdk.Msg
		expErrMsg string
	}{
		{
			name:  "successfully deserialize AuthorizeTx",
			event: testtypes.TestAuthorizeEvent1,
			expMsgs: []sdk.Msg{
				&banktypes.MsgSend{
					FromAddress: testtypes.TestFrom1Eth,
					ToAddress:   testtypes.TestTo3,
					Amount: []sdk.Coin{
						{Denom: "ufuel", Amount: amt},
					},
				},
			},
		},
		{
			name: "errors if AuthorizeTx cannot be deserialized",
			event: &sidecartypes.AuthorizeEvent{
				From:    testtypes.TestFrom1,
				Message: []byte(testtypes.TestMessage1), // Fails because the msg should be encoded to bytes using proto
			},
			expErrMsg: "proto: illegal wireType",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			msgs, err := s.App.BridgeKeeper.DeserializeAuthorizeTx(s.App.AppCodec(), tc.event)

			if len(tc.expErrMsg) > 0 {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expErrMsg)
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(tc.expMsgs, msgs)
		})
	}
}

func (s *KeeperTestSuite) TestAuthorizedMessage() {
	testCases := []struct {
		name            string
		messagesAllowed []string
		msg             sdk.Msg
		expResult       bool
	}{
		{
			name:            "returns true if message is authorized (messages allowed is not *)",
			messagesAllowed: []string{"msg1", "msg2", sdk.MsgTypeURL(&banktypes.MsgSend{})},
			msg: &banktypes.MsgSend{
				FromAddress: "addr1",
				ToAddress:   "addr2",
				Amount:      nil,
			},
			expResult: true,
		},
		{
			name:            "returns true if message is authorized (messages allowed is *)",
			messagesAllowed: []string{"*"},
			msg: &banktypes.MsgSend{
				FromAddress: "addr1",
				ToAddress:   "addr2",
				Amount:      nil,
			},
			expResult: true,
		},
		{
			name:            "returns false if message is not authorized",
			messagesAllowed: []string{"msg1", "msg2", "msg3"},
			msg: &banktypes.MsgSend{
				FromAddress: "addr1",
				ToAddress:   "addr2",
				Amount:      nil,
			},
			expResult: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			actualResult := s.App.BridgeKeeper.AuthorizedMessage(tc.messagesAllowed, tc.msg)
			s.Require().Equal(tc.expResult, actualResult)
		})
	}
}
