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
	TestSeqAddr1Str = "fuelsequencer13tch2uhman7dhjjphmx9uwx7kvg2kqfj5y56hsmljlv93pgma5vqyks99k"
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
	TestFrom1Seq  = "fuelsequencer13hfdkxj5aeqzsll569mqreedkafp34ngcsjkqjmpq6prtgv80kcq83gttw"
	TestFrom2Seq  = "fuelsequencer10xafxk7jmjpfeh6394mvysjaazcjhhtegz5hqlsswgurdqq0e75q9r9vmv"
	TestFrom3Seq  = "fuelsequencer1ssymf5jyka89gsjc9famezv2lcsed7uldfq2z8tkmvg9q3etj2qquzra5n"
	TestAmount1   = "100"
	TestAmount2   = "101"
	TestAmount3   = "102"
	TestTo1       = "fuelsequencer1vtfzrk6f4m6kxt6ehyqt9j5su5hvcz5q3dmlsm"
	TestTo2       = ""
	TestTo3       = "fuelsequencer163rsv65t4893t2rz5rmda9sly7lgdlq2jgr36m"
	TestTo4       = "163rsv65t4893t2rz5rmda9sly7lgdlq2jgr36m"
	TestDuration1 = "31536050"
	TestDuration2 = "31536051"
	TestDuration3 = "31536052"
	TestDuration4 = "abc"
	TestDuration5 = "1"
	// TestMessage1 corresponds to a 10ufuel bank send to TestTo3 from TestFrom1Eth. This was generated with the help
	// of utils/proto_serialization_test.go.
	TestMessage1 = "0aae010a1c2f636f736d6f732e62616e6b2e763162657461312e4d736753656e64128d010a486675656c7365717565" +
		"6e63657231336866646b786a356165717a736c6c3536396d71726565646b61667033346e6763736a6b716a6d70713670727467763830" +
		"6b637138336774747712346675656c73657175656e636572313633727376363574343839337432727a35726d646139736c79376c6764" +
		"6c71326a677233366d1a0b0a05756675656c12023130"

	// TestMessage2 corresponds to a 10ufuel bank send to TestTo3 from TestFrom2Eth. This was generated with the help
	// of utils/proto_serialization_test.go.
	TestMessage2 = "0aae010a1c2f636f736d6f732e62616e6b2e763162657461312e4d736753656e64128d010a486675656c73657175656" +
		"e6365723130786166786b376a6d6a70666568363339346d7679736a61617a636a68687465677a3568716c73737767757264717130653" +
		"73571397239766d7612346675656c73657175656e636572313633727376363574343839337432727a35726d646139736c79376c67646" +
		"c71326a677233366d1a0b0a05756675656c12023130"

	// TestMessage4 corresponds to a 10ufuel bank send to TestTo3 from TestFrom3Eth. This was generated with the help
	// of utils/proto_serialization_test.go.
	TestMessage3 = "0aae010a1c2f636f736d6f732e62616e6b2e763162657461312e4d736753656e64128d010a486675656c73657175656" +
		"e636572317373796d66356a796b61383967736a633966616d657a76326c6373656437756c646671327a38746b6d766739713365746a3" +
		"27171757a7261356e12346675656c73657175656e636572313633727376363574343839337432727a35726d646139736c79376c67646" +
		"c71326a677233366d1a0b0a05756675656c12023130"

	// TestMessage4 corresponds to two 10 ufuel bank sends from testtypes.TestFrom3Seq to testtypes.TestTo3. This was
	// generated with the help of utils/proto_serialization_test.go.
	TestMessage4 = "0aae010a1c2f636f736d6f732e62616e6b2e763162657461312e4d736753656e64128d010a486675656c73657175656e6" +
		"36572317373796d66356a796b61383967736a633966616d657a76326c6373656437756c646671327a38746b6d766739713365746a327" +
		"171757a7261356e12346675656c73657175656e636572313633727376363574343839337432727a35726d646139736c79376c67646c7" +
		"1326a677233366d1a0b0a05756675656c120231300aae010a1c2f636f736d6f732e62616e6b2e763162657461312e4d736753656e641" +
		"28d010a486675656c73657175656e636572317373796d66356a796b61383967736a633966616d657a76326c6373656437756c6466713" +
		"27a38746b6d766739713365746a327171757a7261356e12346675656c73657175656e636572313633727376363574343839337432727" +
		"a35726d646139736c79376c67646c71326a677233366d1a0b0a05756675656c12023130"

	// TestMessage5 corresponds to a MsgWithdrawToEthereum of 0 ufuel from testtypes.TestFrom3Seq. This was generated
	// with the help of utils/proto_serialization_test.go.
	TestMessage5 = "0abc010a2b2f6675656c73657175656e6365722e6272696467652e4d73675769746864726177546f457468657265756d" +
		"128c010a486675656c73657175656e636572317373796d66356a796b61383967736a633966616d657a76326c6373656437756c64667" +
		"1327a38746b6d766739713365746a327171757a7261356e12346675656c73657175656e636572313633727376363574343839337432" +
		"727a35726d646139736c79376c67646c71326a677233366d1a0a0a05756675656c120130"

	// TestMessage6 corresponds to one 10 ufuel and another 1000000ufuel bank send from testtypes.TestFrom3Seq to
	// testtypes.TestTo3. This was generated with the help of utils/proto_serialization_test.go.
	TestMessage6 = "0aae010a1c2f636f736d6f732e62616e6b2e763162657461312e4d736753656e64128d010a486675656c73657175656e" +
		"636572317373796d66356a796b61383967736a633966616d657a76326c6373656437756c646671327a38746b6d766739713365746a3" +
		"27171757a7261356e12346675656c73657175656e636572313633727376363574343839337432727a35726d646139736c79376c6764" +
		"6c71326a677233366d1a0b0a05756675656c120231300ab4010a1c2f636f736d6f732e62616e6b2e763162657461312e4d736753656" +
		"e641293010a486675656c73657175656e636572317373796d66356a796b61383967736a633966616d657a76326c6373656437756c64" +
		"6671327a38746b6d766739713365746a327171757a7261356e12346675656c73657175656e636572313633727376363574343839337" +
		"432727a35726d646139736c79376c67646c71326a677233366d1a110a05756675656c12083130303030303030"

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

	TestEvent1 = testutils.MustGetSidecarEventFromParsedEvent(TestSendToSequencerEvent3)
	TestEvent2 = testutils.MustGetSidecarEventFromParsedEvent(TestAuthorizeEvent3)
	TestEvent3 = testutils.MustGetSidecarEventFromParsedEvent(TestSendToSequencerEvent2)
	TestEvent4 = testutils.MustGetSidecarEventFromParsedEvent(TestSendToSequencerEvent4)
	TestEvent5 = testutils.MustGetSidecarEventFromParsedEvent(TestSendToSequencerEvent5)
	TestEvent6 = testutils.MustGetSidecarEventFromParsedEvent(TestSendToSequencerEvent6)
	TestEvent7 = testutils.MustGetSidecarEventFromParsedEvent(TestSendToSequencerEvent7)
	TestEvent8 = testutils.MustGetSidecarEventFromParsedEvent(TestSendToSequencerEvent8)
	TestEvent9 = testutils.MustGetSidecarEventFromParsedEvent(TestSendToSequencerEvent9)

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
