package app_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/fuel-infrastructure/fuel-sequencer/app/apptesting"
)

type AppTestSuite struct {
	apptesting.KeeperTestHelper
}

func (s *AppTestSuite) SetupTest() {
	s.Setup()
}

func TestAppTestSuite(t *testing.T) {
	suite.Run(t, new(AppTestSuite))
}

func (s *AppTestSuite) TestPowerReductionIsAsExpected() {

	expectPowerReduction, ok := sdkmath.NewIntFromString("1000000000000000000")
	s.Require().True(ok)

	s.Require().True(expectPowerReduction.Equal(s.App.StakingKeeper.PowerReduction(s.Ctx())))
	s.Require().True(expectPowerReduction.Equal(sdk.DefaultPowerReduction))
}
