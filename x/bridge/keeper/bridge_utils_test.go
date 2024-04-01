package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
