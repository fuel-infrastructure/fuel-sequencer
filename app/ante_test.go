package app_test

import (
	"math"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/fuel-infrastructure/fuel-sequencer/app"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// getTrackedAnteHandler returns an NoOp AnteHandler that tracks whether it was called.
func getTrackedAnteHandler(called *bool) sdk.AnteHandler {
	return func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		if *called {
			panic("tracked ante handler already called")
		}
		*called = true
		return ctx, nil
	}
}

func (s *AppTestSuite) TestInjectedTxsDecorator_AnteHandle_ExecMode() {

	dummyTx := sdk.Tx(nil)         // this will not be processed, and if it is, it would cause issues
	nonZeroBlockHeight := int64(1) // ensures the AnteHandler runs as if it's not at genesis

	testCases := []struct {
		name                        string
		execMode                    sdk.ExecMode
		expectNextAnteHandlerCalled bool
	}{
		{
			name:                        "non-finalize exec mode (ExecModeCheck) skips InjectedTxsDecorator",
			execMode:                    sdk.ExecModeCheck,
			expectNextAnteHandlerCalled: true,
		},
		{
			name:                        "non-finalize exec mode (ExecModeReCheck) skips InjectedTxsDecorator",
			execMode:                    sdk.ExecModeReCheck,
			expectNextAnteHandlerCalled: true,
		},
		{
			name:                        "non-finalize exec mode (ExecModeSimulate) skips InjectedTxsDecorator",
			execMode:                    sdk.ExecModeSimulate,
			expectNextAnteHandlerCalled: true,
		},
		{
			name:                        "non-finalize exec mode (ExecModePrepareProposal) skips InjectedTxsDecorator",
			execMode:                    sdk.ExecModePrepareProposal,
			expectNextAnteHandlerCalled: true,
		},
		{
			name:                        "non-finalize exec mode (ExecModeProcessProposal) skips InjectedTxsDecorator",
			execMode:                    sdk.ExecModeProcessProposal,
			expectNextAnteHandlerCalled: true,
		},
		{
			name:                        "non-finalize exec mode (ExecModeVoteExtension) skips InjectedTxsDecorator",
			execMode:                    sdk.ExecModeVoteExtension,
			expectNextAnteHandlerCalled: true,
		},
		{
			name:                        "non-finalize exec mode (ExecModeVerifyVoteExtension) skips InjectedTxsDecorator",
			execMode:                    sdk.ExecModeVerifyVoteExtension,
			expectNextAnteHandlerCalled: true,
		},
		{
			name:                        "finalize exec mode (ExecModeFinalize) stops ante-handling due to no index set",
			execMode:                    sdk.ExecModeFinalize,
			expectNextAnteHandlerCalled: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Index does not exist
			_, found := s.App.BridgeKeeper.GetIndex(s.Ctx())
			s.Require().False(found)

			anteHandlerCalled := false
			trackedAnteHandler := getTrackedAnteHandler(&anteHandlerCalled)

			decorator := app.NewInjectedTxsDecorator(s.App.BridgeKeeper)
			ctx := s.Ctx().WithExecMode(tc.execMode).WithBlockHeight(nonZeroBlockHeight)

			_, err := decorator.AnteHandle(ctx, dummyTx, false, trackedAnteHandler)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectNextAnteHandlerCalled, anteHandlerCalled)

			// Index still does not exist
			_, found = s.App.BridgeKeeper.GetIndex(s.Ctx())
			s.Require().False(found)
		})
	}
}

func (s *AppTestSuite) TestInjectedTxsDecorator_AnteHandle_BlockHeight() {

	dummyTx := sdk.Tx(nil)                   // this will not be processed, and if it is, it would cause issues
	execModeFinalize := sdk.ExecModeFinalize // ensures the AnteHandler runs as if it's at finalization

	testCases := []struct {
		name                        string
		blockHeight                 int64
		expectNextAnteHandlerCalled bool
	}{
		{
			name:                        "zero height (genesis) skips InjectedTxsDecorator",
			blockHeight:                 0,
			expectNextAnteHandlerCalled: true,
		},
		{
			name:                        "non-zero height (1) stops ante-handling due to no index set",
			blockHeight:                 1,
			expectNextAnteHandlerCalled: false,
		},
		{
			name:                        "non-zero height (1000) stops ante-handling due to no index set",
			blockHeight:                 1000,
			expectNextAnteHandlerCalled: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Index does not exist
			_, found := s.App.BridgeKeeper.GetIndex(s.Ctx())
			s.Require().False(found)

			anteHandlerCalled := false
			trackedAnteHandler := getTrackedAnteHandler(&anteHandlerCalled)

			decorator := app.NewInjectedTxsDecorator(s.App.BridgeKeeper)
			ctx := s.Ctx().WithExecMode(execModeFinalize).WithBlockHeight(tc.blockHeight)

			_, err := decorator.AnteHandle(ctx, dummyTx, false, trackedAnteHandler)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectNextAnteHandlerCalled, anteHandlerCalled)

			// Index still does not exist
			_, found = s.App.BridgeKeeper.GetIndex(s.Ctx())
			s.Require().False(found)
		})
	}
}

func (s *AppTestSuite) TestInjectedTxsDecorator_AnteHandle_BehaviourBasedOnIndex() {

	dummyTx := sdk.Tx(nil)                   // this will not be processed, and if it is, it would cause issues
	nonZeroBlockHeight := int64(1)           // ensures the AnteHandler runs as if it's not at genesis
	execModeFinalize := sdk.ExecModeFinalize // ensures the AnteHandler runs as if it's at finalization
	gasLimit := storetypes.Gas(1000)

	testCases := []struct {
		name                        string
		setIndex                    *bridgetypes.Index
		expectIndex                 *bridgetypes.Index
		expectNextAnteHandlerCalled bool
		expectInfiniteGasMetersSet  bool
	}{
		{
			name:                        "no index set => ante-handling stops since we need to process a MsgIndex",
			setIndex:                    nil,
			expectIndex:                 nil,
			expectNextAnteHandlerCalled: false,
			expectInfiniteGasMetersSet:  true, // infinite
		},
		{
			name: "no injected txs seen => injected txs count updated",
			setIndex: &bridgetypes.Index{
				NumInjectedTxsTotal: 5,
				NumInjectedTxsAnte:  0, // no injected txs seen
			},
			expectIndex: &bridgetypes.Index{
				NumInjectedTxsTotal: 5,
				NumInjectedTxsAnte:  1, // incremented
			},
			expectNextAnteHandlerCalled: false,
			expectInfiniteGasMetersSet:  true, // infinite
		},
		{
			name: "not all injected txs seen => injected txs count updated",
			setIndex: &bridgetypes.Index{
				NumInjectedTxsTotal: 5,
				NumInjectedTxsAnte:  1, // some injected txs seen
			},
			expectIndex: &bridgetypes.Index{
				NumInjectedTxsTotal: 5,
				NumInjectedTxsAnte:  2, // incremented
			},
			expectNextAnteHandlerCalled: false,
			expectInfiniteGasMetersSet:  true, // infinite
		},
		{
			name: "all injected txs seen => go to next ante handler to process user transactions",
			setIndex: &bridgetypes.Index{
				NumInjectedTxsTotal: 5,
				NumInjectedTxsAnte:  5,
			},
			expectIndex: &bridgetypes.Index{
				NumInjectedTxsTotal: 5,
				NumInjectedTxsAnte:  5,
			},
			expectNextAnteHandlerCalled: true,
			expectInfiniteGasMetersSet:  false, // finite
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Set up gas meters
			finiteGasMeter := storetypes.NewGasMeter(gasLimit)
			finiteBlockGasMeter := storetypes.NewGasMeter(gasLimit)

			// Index does not exist
			_, found := s.App.BridgeKeeper.GetIndex(s.Ctx())
			s.Require().False(found)

			anteHandlerCalled := false
			trackedAnteHandler := getTrackedAnteHandler(&anteHandlerCalled)

			// Set index
			if tc.setIndex != nil {
				s.App.BridgeKeeper.SetIndex(s.Ctx(), *tc.setIndex)
			}

			decorator := app.NewInjectedTxsDecorator(s.App.BridgeKeeper)
			ctx := s.Ctx().
				WithExecMode(execModeFinalize).
				WithBlockHeight(nonZeroBlockHeight).
				WithGasMeter(finiteGasMeter).
				WithBlockGasMeter(finiteBlockGasMeter)

			updatedCtx, err := decorator.AnteHandle(ctx, dummyTx, false, trackedAnteHandler)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectNextAnteHandlerCalled, anteHandlerCalled)

			// Check index updates if any
			index, found := s.App.BridgeKeeper.GetIndex(s.Ctx())
			if tc.expectIndex != nil {
				s.Require().True(found)
				s.Require().EqualValues(*tc.expectIndex, index)
			} else {
				s.Require().False(found)
			}

			// Check gas meter
			if tc.expectInfiniteGasMetersSet {
				// Some gas from the transaction's gas meter has been consumed, even though the meter is infinite
				s.Require().EqualValues(storetypes.Gas(math.MaxUint64), updatedCtx.GasMeter().Limit())
				s.Require().NotZero(updatedCtx.GasMeter().GasConsumed())
				// Block gas meter is an infinite gas meter with no gas consumed yet (this is not set in AnteHandler)
				s.Require().EqualValues(storetypes.NewInfiniteGasMeter(), updatedCtx.BlockGasMeter())
			} else {
				s.Require().EqualValues(finiteGasMeter, updatedCtx.GasMeter())
				s.Require().EqualValues(finiteBlockGasMeter, updatedCtx.BlockGasMeter())
			}
		})
	}
}
