package types

import (
	sdk "cosmossdk.io/math"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	_ "github.com/fuel-infrastructure/fuel-sequencer/app/apptesting" // Required to load the right config for testing
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	testutils "github.com/fuel-infrastructure/fuel-sequencer/testutil"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

var (
	TestGovernanceAddress = authtypes.NewModuleAddress(govtypes.ModuleName).String()
	TestSupplyDeltaPeriod = uint64(100)
	TestLastEthereumNonce = sdk.NewInt(50)
	TestLastSupply        = sdk.NewInt(100000000)
	TestDelta             = sdk.NewInt(5000000)
	TestOffset            = sdk.NewInt(-2000000)
	TestSupplyDeltaInfo   = bridgetypes.SupplyDeltaInfo{
		LastSupply: TestLastSupply,
		Delta:      TestDelta,
		Offset:     TestOffset,
	}
	TestFrom1                 = "0x000000000000000000000000f39fd6e51aad88f6f4ce6ab8827279cfffb92266"
	TestFrom2                 = "0x0000000000000000000000007E5F4552091A69125d5DfCb7b8C2659029395Bdf"
	TestFrom3                 = "0x000000000000000000000000D1220A0cf47c7B9Be7A2E6BA89F429762e7b9aDb"
	TestAmount1               = "100"
	TestAmount2               = "101"
	TestAmount3               = "102"
	TestTo1                   = "0x4E9ce36E442e55EcD9025B9a6E0D88485d628A67"
	TestTo2                   = "0x0D8775F648430679A709E98d2b0Cb6250d2887EF"
	TestTo3                   = "0x2E645469f354BB4F5c8a05B3b30A929361cf77eC"
	TestDuration1             = "50"
	TestDuration2             = "51"
	TestDuration3             = "52"
	TestMessage1              = "26B5A0378EBB14470BD99C6489279259F8E80E5BD30E4CC84D8385EC334CD936"
	TestMessage2              = "7A1E4C2586F28D5C1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF"
	TestMessage3              = "B23F8D4567E89ABCDEF1234567890ABCDEF1234567890ABCDEF1234567890123"
	TestSendToSequencerEvent1 = &sidecartypes.SendToSequencerEvent{
		From:     TestFrom1,
		Amount:   TestAmount1,
		To:       TestTo1,
		Duration: TestDuration1,
	}
	TestSendToSequencerEvent2 = &sidecartypes.SendToSequencerEvent{
		From:     TestFrom2,
		Amount:   TestAmount2,
		To:       TestTo2,
		Duration: TestDuration2,
	}
	TestSendToSequencerEvent3 = &sidecartypes.SendToSequencerEvent{
		From:     TestFrom3,
		Amount:   TestAmount3,
		To:       TestTo3,
		Duration: TestDuration3,
	}
	TestAuthorizeEvent1 = &sidecartypes.AuthorizeEvent{
		From:    TestFrom1,
		Message: testutils.MustHexDecodeString(TestMessage1),
	}
	TestAuthorizeEvent2 = &sidecartypes.AuthorizeEvent{
		From:    TestFrom2,
		Message: testutils.MustHexDecodeString(TestMessage2),
	}
	TestAuthorizeEvent3 = &sidecartypes.AuthorizeEvent{
		From:    TestFrom3,
		Message: testutils.MustHexDecodeString(TestMessage3),
	}
	TestEvent1      = testutils.MustGetSidecarEventFromParsedEvent(TestSendToSequencerEvent3)
	TestEvent2      = testutils.MustGetSidecarEventFromParsedEvent(TestAuthorizeEvent3)
	TestEvent3      = testutils.MustGetSidecarEventFromParsedEvent(TestSendToSequencerEvent2)
	TestEvents      = []*sidecartypes.Event{TestEvent1, TestEvent2, TestEvent3}
	TestEthEventsTx = &bridgetypes.EthEventsTx{
		Events:           TestEvents,
		AdvanceSequencer: true,
		NewEthereumBlock: true,
	}
	TestSidecarResponse = &sidecartypes.QueryBlockEventsResponse{Events: TestEvents}
)
