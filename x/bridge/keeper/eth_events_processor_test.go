package keeper_test

import (
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	testtypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (s *KeeperTestSuite) TestProcessEthereumEventsSendToSequencerEvent() {

	blockTime, _ := time.Parse(time.DateOnly, "2024-01-01")
	govAddr := s.App.AccountKeeper.GetModuleAddress(govtypes.ModuleName)

	// These accounts correspond to the from and to addresses of the bank.MsgSend to be executed via the AuthorizeEvent
	fromAccOne, _ := types.GenerateSequencerAddressFromEthereumAddress(testtypes.TestFrom3)
	toAccOne := sdk.MustAccAddressFromBech32(testtypes.TestTo3)

	fromAccTwo, _ := types.GenerateSequencerAddressFromEthereumAddress(testtypes.TestFrom2)

	testCases := []struct {
		name           string
		ethEventsTx    *types.EthEventsTx
		toAcc          *sdk.AccAddress
		fromAcc        *sdk.AccAddress
		expFromBalance sdkmath.Int
		expToBalance   sdkmath.Int
		expSupplyDelta *types.SupplyDeltaInfo
		expGovBal      sdkmath.Int
	}{
		{
			name: "successful - send to sequencer - mint to To address",
			ethEventsTx: &types.EthEventsTx{
				Events: []*sidecartypes.Event{
					testtypes.TestEvent1,
				},
				AdvanceSequencer: true,
				NewEthereumBlock: true,
				BlockNumber:      sdkmath.OneInt(),
			},
			fromAcc:        &fromAccOne,
			toAcc:          &toAccOne,
			expFromBalance: sdkmath.NewInt(0),
			expToBalance:   sdkmath.NewInt(102),
			expSupplyDelta: &types.SupplyDeltaInfo{
				Offset: sdkmath.NewInt(-102),
			},
			expGovBal: sdkmath.ZeroInt(),
		},
		{
			name: "successful - send to sequencer - mint to To address twice",
			ethEventsTx: &types.EthEventsTx{
				Events: []*sidecartypes.Event{
					testtypes.TestEvent1,
					testtypes.TestEvent1,
				},
				AdvanceSequencer: true,
				NewEthereumBlock: true,
				BlockNumber:      sdkmath.OneInt(),
			},
			fromAcc:        &fromAccOne,
			toAcc:          &toAccOne,
			expFromBalance: sdkmath.NewInt(0),
			expToBalance:   sdkmath.NewInt(204),
			expSupplyDelta: &types.SupplyDeltaInfo{
				Offset: sdkmath.NewInt(-204),
			},
			expGovBal: sdkmath.ZeroInt(),
		},
		{
			name: "successful - send to sequencer - mint to From address",
			ethEventsTx: &types.EthEventsTx{
				Events: []*sidecartypes.Event{
					testtypes.TestEvent3,
				},
				AdvanceSequencer: true,
				NewEthereumBlock: true,
				BlockNumber:      sdkmath.OneInt(),
			},
			fromAcc:        &fromAccTwo,
			toAcc:          nil,
			expFromBalance: sdkmath.NewInt(101),
			expToBalance:   sdkmath.NewInt(0),
			expSupplyDelta: &types.SupplyDeltaInfo{
				Offset: sdkmath.NewInt(-101),
			},
			expGovBal: sdkmath.ZeroInt(),
		},
		{
			name: "successful - send to sequencer - mint to From address twice",
			ethEventsTx: &types.EthEventsTx{
				Events: []*sidecartypes.Event{
					testtypes.TestEvent3,
					testtypes.TestEvent3,
				},
				AdvanceSequencer: true,
				NewEthereumBlock: true,
				BlockNumber:      sdkmath.OneInt(),
			},
			fromAcc:        &fromAccTwo,
			toAcc:          nil,
			expFromBalance: sdkmath.NewInt(202),
			expToBalance:   sdkmath.NewInt(0),
			expSupplyDelta: &types.SupplyDeltaInfo{
				Offset: sdkmath.NewInt(-202),
			},
			expGovBal: sdkmath.ZeroInt(),
		},
		{
			name: "failure - send to sequencer - duration failed to parse - mint to governance",
			ethEventsTx: &types.EthEventsTx{
				Events: []*sidecartypes.Event{
					testtypes.TestEvent4,
				},
				AdvanceSequencer: true,
				NewEthereumBlock: true,
				BlockNumber:      sdkmath.OneInt(),
			},
			fromAcc:        &fromAccOne,
			toAcc:          &toAccOne,
			expFromBalance: sdkmath.NewInt(0),
			expToBalance:   sdkmath.NewInt(0),
			expSupplyDelta: &types.SupplyDeltaInfo{
				Offset: sdkmath.NewInt(-102),
			},
			expGovBal: sdkmath.NewInt(102),
		},
		{
			name: "failure - send to sequencer - from address failed to parse - mint to governance",
			ethEventsTx: &types.EthEventsTx{
				Events: []*sidecartypes.Event{
					testtypes.TestEvent5,
				},
				AdvanceSequencer: true,
				NewEthereumBlock: true,
				BlockNumber:      sdkmath.OneInt(),
			},
			fromAcc:        &fromAccOne,
			toAcc:          &toAccOne,
			expFromBalance: sdkmath.NewInt(0),
			expToBalance:   sdkmath.NewInt(0),
			expSupplyDelta: &types.SupplyDeltaInfo{
				Offset: sdkmath.NewInt(-102),
			},
			expGovBal: sdkmath.NewInt(102),
		},
		{
			name: "failure - send to sequencer - bad vesting duration - mint to governance",
			ethEventsTx: &types.EthEventsTx{
				Events: []*sidecartypes.Event{
					testtypes.TestEvent6,
				},
				AdvanceSequencer: true,
				NewEthereumBlock: true,
				BlockNumber:      sdkmath.OneInt(),
			},
			fromAcc:        &fromAccOne,
			toAcc:          nil,
			expFromBalance: sdkmath.NewInt(0),
			expToBalance:   sdkmath.NewInt(0),
			expSupplyDelta: &types.SupplyDeltaInfo{
				Offset: sdkmath.NewInt(-102),
			},
			expGovBal: sdkmath.NewInt(102),
		},
		{
			name: "failure - send to sequencer - bad to bech32 address - mint to governance",
			ethEventsTx: &types.EthEventsTx{
				Events: []*sidecartypes.Event{
					testtypes.TestEvent7,
				},
				AdvanceSequencer: true,
				NewEthereumBlock: true,
				BlockNumber:      sdkmath.OneInt(),
			},
			fromAcc:        &fromAccOne,
			toAcc:          &toAccOne,
			expFromBalance: sdkmath.NewInt(0),
			expToBalance:   sdkmath.NewInt(0),
			expSupplyDelta: &types.SupplyDeltaInfo{
				Offset: sdkmath.NewInt(-102),
			},
			expGovBal: sdkmath.NewInt(102),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Set LastEthereumBlockSynced to EthEventsTxs' block height
			s.App.BridgeKeeper.SetLastEthereumBlockSynced(s.Ctx(), tc.ethEventsTx.BlockNumber)

			// Authorize bank.MsgSend on Sequencer
			err := s.App.BridgeKeeper.SetParams(s.Ctx(), types.Params{
				BridgeDenom:              "ufuel",
				AuthorizeMessagesAllowed: []string{"*"},
				VestingStartTime:         blockTime,
			})
			s.Require().NoError(err)

			// Set EthEventsTx
			s.App.BridgeKeeper.SetEthEventsTx(s.Ctx(), *tc.ethEventsTx)
			_, found := s.App.BridgeKeeper.GetEthEventsTx(s.Ctx(), tc.ethEventsTx.BlockNumber.Uint64())
			s.Require().True(found)

			s.App.BridgeKeeper.ProcessEthereumEvents(s.Ctx())

			// Make sure that EthEventsTx has been removed
			_, found = s.App.BridgeKeeper.GetEthEventsTx(s.Ctx(), tc.ethEventsTx.BlockNumber.Uint64())
			s.Require().False(found)

			// Confirm that the balances were changed as specified. This indicates that the AuthorizedEvents were
			// executed successfully
			if tc.fromAcc != nil {
				actualFromBalance := s.App.BankKeeper.GetBalance(s.Ctx(), *tc.fromAcc, "ufuel")
				s.Require().Equal(tc.expFromBalance, actualFromBalance.Amount)
			}
			if tc.toAcc != nil {
				actualToBalance := s.App.BankKeeper.GetBalance(s.Ctx(), *tc.toAcc, "ufuel")
				s.Require().Equal(tc.expToBalance, actualToBalance.Amount)
			}

			// Verify the supply delta info was updated
			if tc.expSupplyDelta != nil {
				actualSupplyDelta := s.App.BridgeKeeper.MustGetSupplyDeltaInfo(s.Ctx())
				s.Require().Equal(tc.expSupplyDelta.Offset, actualSupplyDelta.Offset)
			}
			// Verify the governance address balance is as expected
			actualGovBal := s.App.BankKeeper.GetBalance(s.Ctx(), govAddr, "ufuel")
			s.Require().Equal(tc.expGovBal, actualGovBal.Amount)
		})
	}
}

func (s *KeeperTestSuite) TestProcessEthereumEventsSendToSequencerEvent_AmountParseFailure() {
	blockTime, _ := time.Parse(time.RFC3339, "2024-01-01T00:00:00Z")
	govAddr := s.App.AccountKeeper.GetModuleAddress(govtypes.ModuleName)

	ethEventsTx := &types.EthEventsTx{
		Events: []*sidecartypes.Event{
			testtypes.TestEvent8,
		},
		AdvanceSequencer: true,
		NewEthereumBlock: true,
		BlockNumber:      sdkmath.OneInt(),
	}
	s.SetupTest() // Reset the test suite

	// Set initial conditions: LastEthereumBlockSynced and Params
	s.App.BridgeKeeper.SetLastEthereumBlockSynced(s.Ctx(), ethEventsTx.BlockNumber)

	// Authorize bank.MsgSend on Sequencer
	err := s.App.BridgeKeeper.SetParams(s.Ctx(), types.Params{
		BridgeDenom:              "ufuel",
		AuthorizeMessagesAllowed: []string{"*"},
		VestingStartTime:         blockTime,
	})
	s.Require().NoError(err)

	// Set the malformed EthEventsTx
	s.App.BridgeKeeper.SetEthEventsTx(s.Ctx(), *ethEventsTx)
	_, found := s.App.BridgeKeeper.GetEthEventsTx(s.Ctx(), ethEventsTx.BlockNumber.Uint64())
	s.Require().True(found)

	// Trigger the processing of the Ethereum events
	s.Require().Panics(func() {
		s.App.BridgeKeeper.ProcessEthereumEvents(s.Ctx())
	})

	// Verify the governance address balance is as expected
	actualGovBal := s.App.BankKeeper.GetBalance(s.Ctx(), govAddr, "ufuel")
	s.Require().Equal(sdkmath.ZeroInt(), actualGovBal.Amount)
}
