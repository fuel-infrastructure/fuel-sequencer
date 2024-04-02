package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	evidencetypes "cosmossdk.io/x/evidence/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	authz "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	testtypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	sequencertypes "github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/types"
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

	// address to block
	addressesToBlock := []string{}

	// Block module addresses
	modulesToBlock := []string{
		authtypes.ModuleName,
		authtypes.FeeCollectorName,
		vestingtypes.ModuleName,
		banktypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		"tx",
		genutiltypes.ModuleName,
		authz.ModuleName,
		upgradetypes.ModuleName,
		distrtypes.ModuleName,
		evidencetypes.ModuleName,
		minttypes.ModuleName,
		govtypes.ModuleName,
		crisistypes.ModuleName,
		consensustypes.ModuleName,
		bridgetypes.ModuleName,
		sequencertypes.ModuleName,
	}

	// Block all of the above module addresses
	for _, moduleName := range modulesToBlock {
		addr := s.App.AccountKeeper.GetModuleAddress(moduleName)
		if addr != nil {
			addressesToBlock = append(addressesToBlock, addr.String())
		}
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
			name:                     "successfully retreived blocked addresses",
			paramsBlockedAddresses:   []string{fromAccOne.String()},
			notBlockedAddresses:      []string{fromAccTwo.String()},
			expectedBlockedAddresses: addressesToBlock,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			ctx := s.Ctx()

			vals, err := s.App.StakingKeeper.GetAllValidators(ctx)
			s.Require().NoError(err)

			// Verify that there is at least one validator
			s.Require().GreaterOrEqual(len(vals), 1)

			for _, validator := range vals {
				tc.expectedBlockedAddresses = append(tc.expectedBlockedAddresses, validator.GetOperator())
			}

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
