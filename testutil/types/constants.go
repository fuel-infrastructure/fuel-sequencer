package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	_ "github.com/fuel-infrastructure/fuel-sequencer/app/apptesting" // Required to load the right config for testing
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	testutils "github.com/fuel-infrastructure/fuel-sequencer/testutil"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

var (
	// TestEthAddr1Str maps to TestSeqAddr1Str deterministically
	TestEthAddr1Str = "0x71C7656EC7ab88b098defB751B7401B5f6d8976F"
	TestSeqAddr1Str = "fuelsequencer1w8rk2mk84wytpxx7ld63kaqpkhmd39m05xlgt4"
	TestSeqAddr1    = sdk.MustAccAddressFromBech32(TestSeqAddr1Str)

	FirstAccountSequence  = uint64(0)
	TestToken             = "token"
	TestGovernanceAddress = authtypes.NewModuleAddress(govtypes.ModuleName).String()
	TestSupplyDeltaPeriod = uint64(100)
	TestLastEthereumNonce = sdkmath.NewInt(50)
	TestLastSupply        = sdkmath.NewInt(100000000)
	TestDelta             = sdkmath.NewInt(5000000)
	TestOffset            = sdkmath.NewInt(-2000000)
	TestSupplyDeltaInfo   = bridgetypes.SupplyDeltaInfo{
		LastSupply: TestLastSupply,
		Delta:      TestDelta,
		Offset:     TestOffset,
	}
	TestFrom1     = "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266"
	TestFrom2     = "0x7E5F4552091A69125d5DfCb7b8C2659029395Bdf"
	TestFrom3     = "0xD1220A0cf47c7B9Be7A2E6BA89F429762e7b9aDb"
	TestFrom4     = "faulty-address"
	TestFrom1Seq  = "fuelsequencer17w0adeg64ky0daxwd2ugyuneellmjgnx5dpmtz"
	TestFrom2Seq  = "fuelsequencer10e0525sfrf53yh2aljmm3sn9jq5njk7lnsk0qn"
	TestFrom3Seq  = "fuelsequencer16y3q5r8503aeheazu6agnapfwch8hxkmajmslm"
	TestAmount1   = "100"
	TestAmount2   = "101"
	TestAmount3   = "102"
	TestTo1       = "0x62d221db49aef5632f59b900b2ca90e52ecc0a80"
	TestTo2       = ""
	TestTo3       = "0xd447066a8ba9cb15a862a0f6de961f27be86fc0a"
	TestTo4       = "163rsv65t4893t2rz5rmda9sly7lgdlq2jgr36m"
	TestDuration1 = "31536050"
	TestDuration2 = "31536051"
	TestDuration3 = "31536052"
	TestDuration4 = "abc"
	TestDuration5 = "1"

	// TestMessage1 corresponds to a 10ufuel bank send to TestTo3 from TestFrom1. This was generated with the help of
	// utils/proto_serialization_test.go.
	TestMessage1 = "0a85010a1c2f636f736d6f732e62616e6b2e763162657461312e4d736753656e6412650a2a30786633396664366535316" +
		"1616438386636663463653661623838323732373963666666623932323636122a3078643434373036366138626139636231356138363" +
		"261306636646539363166323762653836666330611a0b0a05756675656c12023130"

	// TestMessage2 corresponds to a 10ufuel bank send to TestTo3 from TestFrom2. This was generated with the help of
	// utils/proto_serialization_test.go.
	TestMessage2 = "0a85010a1c2f636f736d6f732e62616e6b2e763162657461312e4d736753656e6412650a2a30783745354634353532303" +
		"9314136393132356435446643623762384332363539303239333935426466122a3078643434373036366138626139636231356138363" +
		"261306636646539363166323762653836666330611a0b0a05756675656c12023130"

	// TestMessage3 corresponds to a 10ufuel bank send to TestTo3 from TestFrom3. This was generated with the help of
	// utils/proto_serialization_test.go.
	TestMessage3 = "0a85010a1c2f636f736d6f732e62616e6b2e763162657461312e4d736753656e6412650a2a30784431323230413063663" +
		"4376337423942653741324536424138394634323937363265376239614462122a3078643434373036366138626139636231356138363" +
		"261306636646539363166323762653836666330611a0b0a05756675656c12023130"

	// TestMessage4 corresponds to two 10 ufuel bank sends from TestFrom3 to TestTo3. This was generated with the help
	// of utils/proto_serialization_test.go.
	TestMessage4 = "0a85010a1c2f636f736d6f732e62616e6b2e763162657461312e4d736753656e6412650a2a30784431323230413063663" +
		"4376337423942653741324536424138394634323937363265376239614462122a3078643434373036366138626139636231356138363" +
		"261306636646539363166323762653836666330611a0b0a05756675656c120231300a85010a1c2f636f736d6f732e62616e6b2e76316" +
		"2657461312e4d736753656e6412650a2a307844313232304130636634376337423942653741324536424138394634323937363265376" +
		"239614462122a3078643434373036366138626139636231356138363261306636646539363166323762653836666330611a0b0a05756" +
		"675656c12023130"

	// TestMessage5 corresponds to a MsgWithdrawToEthereum of 0 ufuel from TestFrom3. This was generated with the help
	// of utils/proto_serialization_test.go.
	TestMessage5 = "0a670a2b2f6675656c73657175656e6365722e6272696467652e4d73675769746864726177546f457468657265756d123" +
		"80a2a3078443132323041306366343763374239426537413245364241383946343239373632653762396144621a0a0a05756675656c1" +
		"20130"

	// TestMessage6 corresponds to one 10 ufuel and another 1000000ufuel bank send from TestFrom3 to TestTo3. This was
	// generated with the help of utils/proto_serialization_test.go.
	TestMessage6 = "0a85010a1c2f636f736d6f732e62616e6b2e763162657461312e4d736753656e6412650a2a30784431323230413063663" +
		"4376337423942653741324536424138394634323937363265376239614462122a3078643434373036366138626139636231356138363" +
		"261306636646539363166323762653836666330611a0b0a05756675656c120231300a8a010a1c2f636f736d6f732e62616e6b2e76316" +
		"2657461312e4d736753656e64126a0a2a307844313232304130636634376337423942653741324536424138394634323937363265376" +
		"239614462122a3078643434373036366138626139636231356138363261306636646539363166323762653836666330611a100a05756" +
		"675656c120731303030303030"

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
	TestSendToSequencerEvent4 = &sidecartypes.SendToSequencerEvent{
		From:     TestFrom3,
		Amount:   TestAmount3,
		To:       TestTo3,
		Duration: TestDuration4,
	}
	TestSendToSequencerEvent5 = &sidecartypes.SendToSequencerEvent{
		From:     TestFrom4,
		Amount:   TestAmount3,
		To:       TestTo3,
		Duration: TestDuration3,
	}
	TestSendToSequencerEvent6 = &sidecartypes.SendToSequencerEvent{
		From:     TestFrom3,
		Amount:   TestAmount3,
		To:       TestTo2,
		Duration: TestDuration5,
	}
	TestSendToSequencerEvent7 = &sidecartypes.SendToSequencerEvent{
		From:     TestFrom3,
		Amount:   TestAmount3,
		To:       TestTo4,
		Duration: TestDuration3,
	}
	TestSendToSequencerEvent8 = &sidecartypes.SendToSequencerEvent{
		From:     TestFrom3,
		Amount:   "failed",
		To:       TestTo4,
		Duration: TestDuration3,
	}
	TestSendToSequencerEvent9 = &sidecartypes.SendToSequencerEvent{
		From:     TestFrom2,
		Amount:   TestAmount2,
		To:       "fuelsequencer1vtfzrk6f4m6kxt6ehyqt9j5su5hvcz5q3dmlsm",
		Duration: TestDuration2,
	}
	TestSendToSequencerEvent10 = &sidecartypes.SendToSequencerEvent{
		From:     TestFrom3,
		Amount:   TestAmount1,
		To:       TestFrom3,
		Duration: TestDuration1,
	}
	TestSendToSequencerEvent11 = &sidecartypes.SendToSequencerEvent{
		From:     TestFrom3,
		Amount:   TestAmount1,
		To:       TestFrom3Seq,
		Duration: TestDuration1,
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

	TestEvent1  = testutils.MustGetSidecarEventFromParsedEvent(TestSendToSequencerEvent3)
	TestEvent2  = testutils.MustGetSidecarEventFromParsedEvent(TestAuthorizeEvent3)
	TestEvent3  = testutils.MustGetSidecarEventFromParsedEvent(TestSendToSequencerEvent2)
	TestEvent4  = testutils.MustGetSidecarEventFromParsedEvent(TestSendToSequencerEvent4)
	TestEvent5  = testutils.MustGetSidecarEventFromParsedEvent(TestSendToSequencerEvent5)
	TestEvent6  = testutils.MustGetSidecarEventFromParsedEvent(TestSendToSequencerEvent6)
	TestEvent7  = testutils.MustGetSidecarEventFromParsedEvent(TestSendToSequencerEvent7)
	TestEvent8  = testutils.MustGetSidecarEventFromParsedEvent(TestSendToSequencerEvent8)
	TestEvent9  = testutils.MustGetSidecarEventFromParsedEvent(TestSendToSequencerEvent9)
	TestEvent10 = testutils.MustGetSidecarEventFromParsedEvent(TestSendToSequencerEvent10)
	TestEvent11 = testutils.MustGetSidecarEventFromParsedEvent(TestSendToSequencerEvent11)

	TestEvents          = []*sidecartypes.Event{TestEvent1, TestEvent2, TestEvent3}
	TestEventsDifferent = []*sidecartypes.Event{TestEvent3, TestEvent1, TestEvent2} // jumbled up
	TestEventsReduced   = []*sidecartypes.Event{TestEvent1, TestEvent2}

	TestEthEventsTx = bridgetypes.EthEventsTx{
		Events:           TestEvents,
		AdvanceSequencer: true,
		NewEthereumBlock: true,
		BlockNumber:      1,
	}
	TestEthEventsTxWithDifferentEvents = bridgetypes.EthEventsTx{
		Events:           TestEventsDifferent,
		AdvanceSequencer: true,
		NewEthereumBlock: true,
		BlockNumber:      1,
	}
	TestEthEventsTxReduced = bridgetypes.EthEventsTx{
		Events:           TestEventsReduced,
		AdvanceSequencer: true,
		NewEthereumBlock: true,
		BlockNumber:      1,
	}
	TestEthEventsTxPartial = bridgetypes.EthEventsTx{
		Events:           TestEventsReduced,
		AdvanceSequencer: true,
		NewEthereumBlock: false, // block was partially consumed
		BlockNumber:      1,
	}
	TestEthEventsTxWithoutEvents = bridgetypes.EthEventsTx{
		Events:           []*sidecartypes.Event{},
		AdvanceSequencer: true,
		NewEthereumBlock: true,
		BlockNumber:      1,
	}
	TestEthEventsTxNoNewBlock = bridgetypes.EthEventsTx{
		Events:           []*sidecartypes.Event{},
		AdvanceSequencer: true,
		NewEthereumBlock: false,
		BlockNumber:      1,
	}
	TestEthEventsTxSidecarDelay = bridgetypes.EthEventsTx{
		Events:           []*sidecartypes.Event{},
		AdvanceSequencer: true,
		NewEthereumBlock: false,
		BlockNumber:      1,
	}
	TestEthEventsTxSidecarErr = bridgetypes.EthEventsTx{
		Events:           []*sidecartypes.Event{},
		AdvanceSequencer: false,
		NewEthereumBlock: false,
		BlockNumber:      1,
	}

	TestEmptySidecarResponse   = &sidecartypes.QueryBlockEventsResponse{Events: nil}
	TestSidecarResponse        = &sidecartypes.QueryBlockEventsResponse{Events: TestEvents}
	TestSidecarResponseReduced = &sidecartypes.QueryBlockEventsResponse{Events: TestEventsReduced}
)
