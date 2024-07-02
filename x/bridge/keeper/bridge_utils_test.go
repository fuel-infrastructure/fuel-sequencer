package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/testutil"
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

func (s *KeeperTestSuite) TestGetAllBlockedAddresses() {

	var addressesToBlock []string

	for _, permission := range s.App.AccountKeeper.GetModulePermissions() {
		addrStr, err := s.App.AccountKeeper.AddressCodec().BytesToString(permission.GetAddress())
		if err != nil {
			s.Require().NoError(err)
		}
		addressesToBlock = append(addressesToBlock, addrStr)
	}

	fromAccOne := sdk.MustAccAddressFromBech32(testtypes.TestFrom3Seq)
	addressesToBlock = append(addressesToBlock, fromAccOne.String())

	fromAccTwo := sdk.MustAccAddressFromBech32(testtypes.TestFrom2Seq)

	testCases := []struct {
		name                     string
		paramsBlockedAddresses   []string
		notBlockedAddresses      []string
		expectedBlockedAddresses []string
	}{
		{
			name:                     "successfully retrieved blocked addresses",
			paramsBlockedAddresses:   []string{fromAccOne.String()},
			notBlockedAddresses:      []string{fromAccTwo.String()},
			expectedBlockedAddresses: addressesToBlock,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			ctx := s.Ctx()

			// Create validator which will not get included in the GetLastValidators call
			valPubKey := simtestutil.CreateTestPubKeys(1)[0]
			valAddr := sdk.ValAddress(valPubKey.Address().Bytes())
			addTokens := s.App.StakingKeeper.TokensFromConsensusPower(ctx, 10)
			extraValidator := testutil.NewValidator(s.T(), valAddr, valPubKey)
			extraValidator, _, err := s.App.StakingKeeper.AddValidatorTokensAndShares(ctx, extraValidator, addTokens)
			s.Require().NoError(err)
			s.Require().Equal(addTokens, extraValidator.Tokens)

			vals, err := s.App.StakingKeeper.GetLastValidators(ctx)
			s.Require().NoError(err)

			// Confirm the difference between the 'LastValidators' and the full list of validators.
			allVals, err := s.App.StakingKeeper.GetAllValidators(ctx)
			s.Require().NoError(err)
			s.Require().Greater(len(allVals), len(vals))

			// Verify that there is at least one validator
			s.Require().GreaterOrEqual(len(vals), 1)

			// All account addresses of validators are blocked
			for _, validator := range vals {
				valAddr, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
				s.Require().NoError(err)
				accAddrFromValAddr := sdk.AccAddress(valAddr.Bytes()).String()
				tc.expectedBlockedAddresses = append(tc.expectedBlockedAddresses, accAddrFromValAddr)
			}

			// 'Extra' validator is not blocked
			extraValAddr, err := sdk.ValAddressFromBech32(extraValidator.OperatorAddress)
			s.Require().NoError(err)
			accAddrFromExtraValAddr := sdk.AccAddress(extraValAddr.Bytes()).String()
			tc.notBlockedAddresses = append(tc.notBlockedAddresses, accAddrFromExtraValAddr)

			// The authority address is blocked
			tc.expectedBlockedAddresses = append(tc.expectedBlockedAddresses, s.App.BridgeKeeper.GetAuthority())

			blockedAddresses, err := s.App.BridgeKeeper.GetAllBlockedAddresses(ctx, tc.paramsBlockedAddresses)
			s.Require().NoError(err)

			// Verify that all expected addresses are marked as blocked
			for _, addr := range tc.expectedBlockedAddresses {
				s.Require().True(blockedAddresses[addr], fmt.Sprintf("%s should be blocked", addr))
			}

			// Verify that certain addresses are not blocked
			for _, addr := range tc.notBlockedAddresses {
				s.Require().False(blockedAddresses[addr], fmt.Sprintf("%s should not be blocked", addr))
			}
		})
	}
}
