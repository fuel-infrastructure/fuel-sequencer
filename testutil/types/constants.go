package types

import (
	"time"

	sdkmath "cosmossdk.io/math"
	govtypes "cosmossdk.io/x/gov/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	_ "github.com/fuel-infrastructure/fuel-sequencer/app/apptesting" // Required to load the right config for testing
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	testutils "github.com/fuel-infrastructure/fuel-sequencer/testutil"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	reportstypes "github.com/fuel-infrastructure/fuel-sequencer/x/reports/types"
)

var (
	// TestEthAddr1Str maps to TestSeqAddr1Str deterministically
	TestEthAddr1Str = "0x71C7656EC7ab88b098defB751B7401B5f6d8976F"
	TestSeqAddr1Str = "fuelsequencer1w8rk2mk84wytpxx7ld63kaqpkhmd39m05xlgt4"
	TestSeqAddr2Str = "fuelsequencer1n79wsstpakv0gw2efmruf9x8xs9m4rqfazfu8g"
	TestValAddr1Str = "fuelsequencervaloper1w8rk2mk84wytpxx7ld63kaqpkhmd39m0ky2e6a"
	TestSeqAddr1    = sdk.MustAccAddressFromBech32(TestSeqAddr1Str)
	TestSeqAddr2    = sdk.MustAccAddressFromBech32(TestSeqAddr2Str)

	FirstAccountSequence             = uint64(0)
	TestToken                        = "token"
	TestGovernanceAddress            = authtypes.NewModuleAddress(govtypes.ModuleName).String()
	TestSupplyDeltaPeriod            = uint64(100)
	TestEthereumProxyContractAddress = "0x0165878A594ca255338adfa4d48449f69242Eb8F"
	TestInjectedEventTxMaxBytes      = uint64(20_000_000)
	TestMaxAuthorizeMessages         = uint64(10)
	TestSequencerTxsAllocation       = sdkmath.LegacyMustNewDecFromStr("0.3")
	TestLastEthereumNonce            = sdkmath.NewInt(50)
	TestVestingStartingTime          = time.Now()
	TestBridgeDenomTotalSupply       = sdkmath.NewInt(10_000_000_000)
	TestLastSupply                   = sdkmath.NewInt(100_000_000)
	TestToReport                     = sdkmath.NewInt(5_000_000)
	TestOffset                       = sdkmath.NewInt(-2_000_000)
	TestSupplyDeltaInfo              = bridgetypes.SupplyDeltaInfo{
		LastSupply: TestLastSupply,
		Offset:     TestOffset,
		ToReport:   TestToReport,
	}
	TestFrom1        = "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266"
	TestFrom2        = "0x7E5F4552091A69125d5DfCb7b8C2659029395Bdf"
	TestFrom3        = "0xD1220A0cf47c7B9Be7A2E6BA89F429762e7b9aDb"
	TestFrom4        = "faulty-address"
	TestFrom1Seq     = "fuelsequencer17w0adeg64ky0daxwd2ugyuneellmjgnx5dpmtz"
	TestFrom2Seq     = "fuelsequencer10e0525sfrf53yh2aljmm3sn9jq5njk7lnsk0qn"
	TestFrom3Seq     = "fuelsequencer16y3q5r8503aeheazu6agnapfwch8hxkmajmslm"
	TestFrom1Val     = "fuelsequencervaloper17w0adeg64ky0daxwd2ugyuneellmjgnxk05262"
	TestFrom2Val     = "fuelsequencervaloper10e0525sfrf53yh2aljmm3sn9jq5njk7l3jr73m"
	TestFrom3Val     = "fuelsequencervaloper16y3q5r8503aeheazu6agnapfwch8hxkmlswpwn"
	TestAmount1      = "100"
	TestAmount2      = "101"
	TestAmount3      = "102"
	TestTo1          = "0x62d221db49aef5632f59b900b2ca90e52ecc0a80"
	TestTo2          = "0x0000000000000000000000000000000000000000" // the null Ethereum address
	TestTo3          = "0xd447066a8ba9cb15a862a0f6de961f27be86fc0a"
	TestTo4          = "163rsv65t4893t2rz5rmda9sly7lgdlq2jgr36m"
	TestLockup1      = "31536050"
	TestLockup2      = "31536051"
	TestLockup3      = "31536052"
	TestLockup4      = "abc"
	TestLockup5      = "1"
	ValidSlashEntry1 = reportstypes.SlashEntry{
		ValidatorAddress:          TestFrom1Val,
		DelegatorAddress:          TestFrom1Seq,
		DelegatorSlashAmount:      sdkmath.OneInt(),
		DelegatorBondedBalance:    sdkmath.NewInt(20),
		DelegatorUnbondingBalance: sdkmath.NewInt(10),
	}
	ValidSlashEntry2 = reportstypes.SlashEntry{
		ValidatorAddress:          TestFrom2Val,
		DelegatorAddress:          TestFrom2Seq,
		DelegatorSlashAmount:      sdkmath.NewInt(2),
		DelegatorBondedBalance:    sdkmath.NewInt(30),
		DelegatorUnbondingBalance: sdkmath.NewInt(20),
	}
	ValidSlashEntry3 = reportstypes.SlashEntry{
		ValidatorAddress:          TestFrom3Val,
		DelegatorAddress:          TestFrom3Seq,
		DelegatorSlashAmount:      sdkmath.NewInt(3),
		DelegatorBondedBalance:    sdkmath.NewInt(0),
		DelegatorUnbondingBalance: sdkmath.NewInt(0),
	}
	InvalidSlashEntryValidatorAddressNotValoper = reportstypes.SlashEntry{
		ValidatorAddress:          TestFrom2Seq,
		DelegatorAddress:          TestFrom2Seq,
		DelegatorSlashAmount:      sdkmath.NewInt(2),
		DelegatorBondedBalance:    sdkmath.NewInt(30),
		DelegatorUnbondingBalance: sdkmath.NewInt(20),
	}
	InvalidSlashEntryValidatorAddressNotAccAddress = reportstypes.SlashEntry{
		ValidatorAddress:          TestFrom2Val,
		DelegatorAddress:          TestFrom2Val,
		DelegatorSlashAmount:      sdkmath.NewInt(2),
		DelegatorBondedBalance:    sdkmath.NewInt(30),
		DelegatorUnbondingBalance: sdkmath.NewInt(20),
	}
	InvalidSlashEntryNegativeDelegatorSlashAmount = reportstypes.SlashEntry{
		ValidatorAddress:          TestFrom2Val,
		DelegatorAddress:          TestFrom2Seq,
		DelegatorSlashAmount:      sdkmath.NewInt(-1),
		DelegatorBondedBalance:    sdkmath.NewInt(30),
		DelegatorUnbondingBalance: sdkmath.NewInt(20),
	}
	InvalidSlashEntryZeroDelegatorSlashAmount = reportstypes.SlashEntry{
		ValidatorAddress:          TestFrom2Val,
		DelegatorAddress:          TestFrom2Seq,
		DelegatorSlashAmount:      sdkmath.ZeroInt(),
		DelegatorBondedBalance:    sdkmath.NewInt(30),
		DelegatorUnbondingBalance: sdkmath.NewInt(20),
	}
	InvalidSlashEntryNegativeDelegatorBondedBalance = reportstypes.SlashEntry{
		ValidatorAddress:          TestFrom2Val,
		DelegatorAddress:          TestFrom2Seq,
		DelegatorSlashAmount:      sdkmath.NewInt(2),
		DelegatorBondedBalance:    sdkmath.NewInt(-1),
		DelegatorUnbondingBalance: sdkmath.NewInt(20),
	}
	InvalidSlashEntryNegativeDelegatorUnbondingBalance = reportstypes.SlashEntry{
		ValidatorAddress:          TestFrom2Val,
		DelegatorAddress:          TestFrom2Seq,
		DelegatorSlashAmount:      sdkmath.NewInt(2),
		DelegatorBondedBalance:    sdkmath.NewInt(30),
		DelegatorUnbondingBalance: sdkmath.NewInt(-1),
	}
	InvalidSlashEntryNonUnique = reportstypes.SlashEntry{
		ValidatorAddress:          TestFrom1Val, // Equal to ValidSlashEntry1
		DelegatorAddress:          TestFrom1Seq, // Equal to ValidSlashEntry1
		DelegatorSlashAmount:      sdkmath.NewInt(2),
		DelegatorBondedBalance:    sdkmath.NewInt(30),
		DelegatorUnbondingBalance: sdkmath.NewInt(20),
	}
	ValidSlashReport1 = reportstypes.SlashReport{
		Height:  1,
		Entries: []reportstypes.SlashEntry{ValidSlashEntry1, ValidSlashEntry2},
	}
	ValidSlashReport2 = reportstypes.SlashReport{
		Height:  2,
		Entries: []reportstypes.SlashEntry{ValidSlashEntry1},
	}
	ValidSlashReport3 = reportstypes.SlashReport{
		Height:  3,
		Entries: []reportstypes.SlashEntry{ValidSlashEntry3},
	}
	InvalidSlashReportHeightZero = reportstypes.SlashReport{
		Height:  0,
		Entries: []reportstypes.SlashEntry{ValidSlashEntry1, ValidSlashEntry2},
	}
	InvalidSlashReportEmptyEntries = reportstypes.SlashReport{
		Height:  1,
		Entries: []reportstypes.SlashEntry{},
	}
	InvalidSlashReportNilEntries = reportstypes.SlashReport{
		Height:  1,
		Entries: nil,
	}
	InvalidSlashReportValidatorAddressNotValoper = reportstypes.SlashReport{
		Height:  1,
		Entries: []reportstypes.SlashEntry{ValidSlashEntry1, InvalidSlashEntryValidatorAddressNotValoper},
	}
	InvalidSlashReportDelegatorAddressNotAccAddress = reportstypes.SlashReport{
		Height:  1,
		Entries: []reportstypes.SlashEntry{ValidSlashEntry1, InvalidSlashEntryValidatorAddressNotAccAddress},
	}
	InvalidSlashReportDelegatorSlashAmountNegative = reportstypes.SlashReport{
		Height:  1,
		Entries: []reportstypes.SlashEntry{ValidSlashEntry1, InvalidSlashEntryNegativeDelegatorSlashAmount},
	}
	InvalidSlashReportDelegatorSlashAmountZero = reportstypes.SlashReport{
		Height:  1,
		Entries: []reportstypes.SlashEntry{ValidSlashEntry1, InvalidSlashEntryZeroDelegatorSlashAmount},
	}
	InvalidSlashReportDelegatorBondedBalanceNegative = reportstypes.SlashReport{
		Height:  1,
		Entries: []reportstypes.SlashEntry{ValidSlashEntry1, InvalidSlashEntryNegativeDelegatorBondedBalance},
	}
	InvalidSlashReportDelegatorUnbondingBalanceNegative = reportstypes.SlashReport{
		Height:  1,
		Entries: []reportstypes.SlashEntry{ValidSlashEntry1, InvalidSlashEntryNegativeDelegatorUnbondingBalance},
	}
	InvalidSlashReportNonUniqueSlashEntries = reportstypes.SlashReport{
		Height:  1,
		Entries: []reportstypes.SlashEntry{ValidSlashEntry1, InvalidSlashEntryNonUnique},
	}

	// TestData1 corresponds to a 10ufuel bank send to TestTo3 from TestFrom1. This was generated with the help of
	// scripts/proto_serialization_test.go.
	TestData1 = "0a85010a1c2f636f736d6f732e62616e6b2e763162657461312e4d736753656e6412650a2a30786633396664366535316161" +
		"6438386636663463653661623838323732373963666666623932323636122a3078643434373036366138626139636231356138363261" +
		"306636646539363166323762653836666330611a0b0a05756675656c12023130"

	// TestData2 corresponds to a 10ufuel bank send to TestTo3 from TestFrom2. This was generated with the help of
	// scripts/proto_serialization_test.go.
	TestData2 = "0a85010a1c2f636f736d6f732e62616e6b2e763162657461312e4d736753656e6412650a2a30783745354634353532303931" +
		"4136393132356435446643623762384332363539303239333935426466122a3078643434373036366138626139636231356138363261" +
		"306636646539363166323762653836666330611a0b0a05756675656c12023130"

	// TestData3 corresponds to a 10ufuel bank send to TestTo3 from TestFrom3. This was generated with the help of
	// scripts/proto_serialization_test.go.
	TestData3 = "0a85010a1c2f636f736d6f732e62616e6b2e763162657461312e4d736753656e6412650a2a30784431323230413063663437" +
		"6337423942653741324536424138394634323937363265376239614462122a3078643434373036366138626139636231356138363261" +
		"306636646539363166323762653836666330611a0b0a05756675656c12023130"

	// TestData4 corresponds to two 10 ufuel bank sends from TestFrom3 to TestTo3. This was generated with the help of
	// scripts/proto_serialization_test.go.
	TestData4 = "0a85010a1c2f636f736d6f732e62616e6b2e763162657461312e4d736753656e6412650a2a30784431323230413063663437" +
		"6337423942653741324536424138394634323937363265376239614462122a3078643434373036366138626139636231356138363261" +
		"306636646539363166323762653836666330611a0b0a05756675656c120231300a85010a1c2f636f736d6f732e62616e6b2e76316265" +
		"7461312e4d736753656e6412650a2a307844313232304130636634376337423942653741324536424138394634323937363265376239" +
		"614462122a3078643434373036366138626139636231356138363261306636646539363166323762653836666330611a0b0a05756675" +
		"656c12023130"

	// TestData5 corresponds to a MsgWithdrawToEthereum of 0 ufuel from TestFrom3. This was generated with the help of
	// scripts/proto_serialization_test.go.
	TestData5 = "0a6a0a2e2f6675656c73657175656e6365722e6272696467652e76312e4d73675769746864726177546f457468657265756d" +
		"12380a2a3078443132323041306366343763374239426537413245364241383946343239373632653762396144621a0a0a0575667565" +
		"6c120130"

	// TestData6 corresponds to one 10 ufuel and another 1000000ufuel bank send from TestFrom3 to TestTo3. This was
	// generated with the help of scripts/proto_serialization_test.go.
	TestData6 = "0a85010a1c2f636f736d6f732e62616e6b2e763162657461312e4d736753656e6412650a2a30784431323230413063663437" +
		"6337423942653741324536424138394634323937363265376239614462122a3078643434373036366138626139636231356138363261" +
		"306636646539363166323762653836666330611a0b0a05756675656c120231300a8a010a1c2f636f736d6f732e62616e6b2e76316265" +
		"7461312e4d736753656e64126a0a2a307844313232304130636634376337423942653741324536424138394634323937363265376239" +
		"614462122a3078643434373036366138626139636231356138363261306636646539363166323762653836666330611a100a05756675" +
		"656c120731303030303030"

	TestDepositEvent1 = &sidecartypes.DepositEvent{
		Depositor: TestFrom1,
		Recipient: TestTo1,
		Amount:    TestAmount1,
		Lockup:    TestLockup1,
	}
	TestDepositEvent2 = &sidecartypes.DepositEvent{
		Depositor: TestFrom2,
		Recipient: TestTo2,
		Amount:    TestAmount2,
		Lockup:    TestLockup2,
	}
	TestDepositEvent3 = &sidecartypes.DepositEvent{
		Depositor: TestFrom3,
		Recipient: TestTo3,
		Amount:    TestAmount3,
		Lockup:    TestLockup3,
	}
	TestDepositEvent4 = &sidecartypes.DepositEvent{
		Depositor: TestFrom3,
		Recipient: TestTo3,
		Amount:    TestAmount3,
		Lockup:    TestLockup4,
	}
	TestDepositEvent5 = &sidecartypes.DepositEvent{
		Depositor: TestFrom4,
		Recipient: TestTo3,
		Amount:    TestAmount3,
		Lockup:    TestLockup3,
	}
	TestDepositEvent6 = &sidecartypes.DepositEvent{
		Depositor: TestFrom3,
		Recipient: TestTo2,
		Amount:    TestAmount3,
		Lockup:    TestLockup5,
	}
	TestDepositEvent7 = &sidecartypes.DepositEvent{
		Depositor: TestFrom3,
		Recipient: TestTo4,
		Amount:    TestAmount3,
		Lockup:    TestLockup3,
	}
	TestDepositEvent8 = &sidecartypes.DepositEvent{
		Depositor: TestFrom3,
		Recipient: TestTo4,
		Amount:    "failed",
		Lockup:    TestLockup3,
	}
	TestDepositEvent9 = &sidecartypes.DepositEvent{
		Depositor: TestFrom2,
		Recipient: "fuelsequencer1vtfzrk6f4m6kxt6ehyqt9j5su5hvcz5q3dmlsm",
		Amount:    TestAmount2,
		Lockup:    TestLockup2,
	}
	TestDepositEvent10 = &sidecartypes.DepositEvent{
		Depositor: TestFrom3,
		Recipient: TestFrom3,
		Amount:    TestAmount1,
		Lockup:    TestLockup1,
	}
	TestDepositEvent11 = &sidecartypes.DepositEvent{
		Depositor: TestFrom3,
		Recipient: TestFrom3Seq,
		Amount:    TestAmount1,
		Lockup:    TestLockup1,
	}
	TestDepositEvent12 = &sidecartypes.DepositEvent{
		Depositor: TestFrom1,
		Recipient: TestTo1,
		Amount:    "faulty-amount",
		Lockup:    TestLockup1,
	}
	TestAuthorizeEvent1 = &sidecartypes.AuthorizeEvent{
		Sender: TestFrom1,
		Data:   testutils.MustHexDecodeString(TestData1),
	}
	TestAuthorizeEvent2 = &sidecartypes.AuthorizeEvent{
		Sender: TestFrom2,
		Data:   testutils.MustHexDecodeString(TestData2),
	}
	TestAuthorizeEvent3 = &sidecartypes.AuthorizeEvent{
		Sender: TestFrom3,
		Data:   testutils.MustHexDecodeString(TestData3),
	}
	TestAuthorizeEvent4 = &sidecartypes.AuthorizeEvent{
		Sender: TestFrom3,
		Data:   []byte("some invalid data"),
	}

	TestEvent1 = testutils.MustGetSidecarEventFromParsedEvent(
		TestDepositEvent3, TestEthereumProxyContractAddress,
	)
	TestEvent2 = testutils.MustGetSidecarEventFromParsedEvent(
		TestAuthorizeEvent3, TestEthereumProxyContractAddress,
	)
	TestEvent3 = testutils.MustGetSidecarEventFromParsedEvent(
		TestDepositEvent2, TestEthereumProxyContractAddress,
	)
	TestEvent4 = testutils.MustGetSidecarEventFromParsedEvent(
		TestDepositEvent4, TestEthereumProxyContractAddress,
	)
	TestEvent5 = testutils.MustGetSidecarEventFromParsedEvent(
		TestDepositEvent5, TestEthereumProxyContractAddress,
	)
	TestEvent6 = testutils.MustGetSidecarEventFromParsedEvent(
		TestDepositEvent6, TestEthereumProxyContractAddress,
	)
	TestEvent7 = testutils.MustGetSidecarEventFromParsedEvent(
		TestDepositEvent7, TestEthereumProxyContractAddress,
	)
	TestEvent8 = testutils.MustGetSidecarEventFromParsedEvent(
		TestDepositEvent8, TestEthereumProxyContractAddress,
	)
	TestEvent9 = testutils.MustGetSidecarEventFromParsedEvent(
		TestDepositEvent9, TestEthereumProxyContractAddress,
	)
	TestEvent10 = testutils.MustGetSidecarEventFromParsedEvent(
		TestDepositEvent10, TestEthereumProxyContractAddress,
	)
	TestEvent11 = testutils.MustGetSidecarEventFromParsedEvent(
		TestDepositEvent11, TestEthereumProxyContractAddress,
	)
	TestEvent12 = testutils.MustGetSidecarEventFromParsedEvent(
		TestDepositEvent12, TestEthereumProxyContractAddress,
	)
	TestEvent13 = testutils.MustGetSidecarEventFromParsedEvent(
		TestAuthorizeEvent4, TestEthereumProxyContractAddress,
	)

	// The below event messages are set in the init() function.
	// These serve as convenient access to the event's original messages.
	// We skipped TestEvent2Msg because this is an authorize msg.

	TestEvent1Msg  *bridgetypes.MsgDepositFromEthereum
	TestEvent3Msg  *bridgetypes.MsgDepositFromEthereum
	TestEvent4Msg  *bridgetypes.MsgDepositFromEthereum
	TestEvent5Msg  *bridgetypes.MsgDepositFromEthereum
	TestEvent6Msg  *bridgetypes.MsgDepositFromEthereum
	TestEvent7Msg  *bridgetypes.MsgDepositFromEthereum
	TestEvent8Msg  *bridgetypes.MsgDepositFromEthereum
	TestEvent9Msg  *bridgetypes.MsgDepositFromEthereum
	TestEvent10Msg *bridgetypes.MsgDepositFromEthereum
	TestEvent11Msg *bridgetypes.MsgDepositFromEthereum
	TestEvent12Msg *bridgetypes.MsgDepositFromEthereum

	TestEvents               = []*sidecartypes.Event{TestEvent1, TestEvent2, TestEvent3}
	TestEventsDifferent      = []*sidecartypes.Event{TestEvent3, TestEvent1, TestEvent2} // jumbled up
	TestEventsReduced        = []*sidecartypes.Event{TestEvent1, TestEvent2}
	TestEventsWithFourEvents = []*sidecartypes.Event{
		TestEvent1, TestEvent2, TestEvent3, TestEvent3,
	}

	TestEventsInvalidDeposit   = []*sidecartypes.Event{TestEvent12}
	TestEventsInvalidAuthorize = []*sidecartypes.Event{TestEvent13}

	TestEventsDepositOnly   = []*sidecartypes.Event{TestEvent1}
	TestEventsAuthorizeOnly = []*sidecartypes.Event{TestEvent2}

	TestMsgSupplyDelta = &bridgetypes.MsgSupplyDelta{
		Authority: TestGovernanceAddress,
	}

	TestMsgDepositFromEthereum = &bridgetypes.MsgDepositFromEthereum{
		Authority: TestGovernanceAddress,
		Depositor: TestFrom1,
		Recipient: TestTo1,
		Amount:    TestAmount1,
		Lockup:    TestLockup1,
	}

	TestMsgIndex = TestMsgIndexWithEvents{
		MsgIndex: &bridgetypes.MsgIndex{
			Authority:           TestGovernanceAddress,
			NumInjectedEventTxs: uint64(len(TestEvents)),
			NewEthereumBlock:    true,
			BlockNumber:         1,
		},
		Events: TestEvents,
	}

	TestMsgIndexWithDifferentEvents = TestMsgIndexWithEvents{
		MsgIndex: &bridgetypes.MsgIndex{
			Authority:           TestGovernanceAddress,
			NumInjectedEventTxs: uint64(len(TestEventsDifferent)),
			NewEthereumBlock:    true,
			BlockNumber:         1,
		},
		Events: TestEventsDifferent,
	}

	TestMsgIndexWithFourEvents = TestMsgIndexWithEvents{
		MsgIndex: &bridgetypes.MsgIndex{
			Authority:           TestGovernanceAddress,
			NumInjectedEventTxs: uint64(len(TestEventsWithFourEvents)),
			NewEthereumBlock:    true,
			BlockNumber:         1,
		},
		Events: TestEventsWithFourEvents,
	}

	TestMsgIndexReduced = TestMsgIndexWithEvents{
		MsgIndex: &bridgetypes.MsgIndex{
			Authority:           TestGovernanceAddress,
			NumInjectedEventTxs: uint64(len(TestEventsReduced)),
			NewEthereumBlock:    true,
			BlockNumber:         1,
		},
		Events: TestEventsReduced,
	}

	TestMsgIndexIncorrectAuthority = TestMsgIndexWithEvents{
		MsgIndex: &bridgetypes.MsgIndex{
			Authority:           TestSeqAddr1Str, // not the governance address
			NumInjectedEventTxs: uint64(len(TestEvents)),
			NewEthereumBlock:    true,
			BlockNumber:         1,
		},
		Events: TestEvents,
	}

	TestMsgIndexWithNoNewEthBlock = TestMsgIndexWithEvents{
		MsgIndex: &bridgetypes.MsgIndex{
			Authority:           TestGovernanceAddress,
			NumInjectedEventTxs: uint64(len(TestEvents)),
			NewEthereumBlock:    false, // no new block
			BlockNumber:         1,
		},
		Events: TestEvents,
	}

	TestMsgIndexWithDiffBlockNumber = TestMsgIndexWithEvents{
		MsgIndex: &bridgetypes.MsgIndex{
			Authority:           TestGovernanceAddress,
			NumInjectedEventTxs: uint64(len(TestEvents)),
			NewEthereumBlock:    true,
			BlockNumber:         100, // not 1
		},
		Events: TestEvents,
	}

	TestMsgIndexPartial = TestMsgIndexWithEvents{
		MsgIndex: &bridgetypes.MsgIndex{
			Authority:           TestGovernanceAddress,
			NumInjectedEventTxs: uint64(len(TestEventsReduced)),
			NewEthereumBlock:    false, // block was partially consumed
			BlockNumber:         1,
		},
		Events: TestEventsReduced,
	}
	TestMsgIndexPartial2 = &TestMsgIndexWithEvents{
		MsgIndex: &bridgetypes.MsgIndex{
			Authority:           TestGovernanceAddress,
			NumInjectedEventTxs: uint64(len(TestEvents)),
			NewEthereumBlock:    false,
			BlockNumber:         1,
		},
		Events: TestEvents,
	}

	TestMsgIndexWithoutEvents = TestMsgIndexWithEvents{
		MsgIndex: &bridgetypes.MsgIndex{
			Authority:           TestGovernanceAddress,
			NumInjectedEventTxs: 0,
			NewEthereumBlock:    true,
			BlockNumber:         1,
		},
		Events: nil,
	}

	TestMsgIndexNoNewBlock = TestMsgIndexWithEvents{
		MsgIndex: &bridgetypes.MsgIndex{
			Authority:           TestGovernanceAddress,
			NumInjectedEventTxs: 0,
			NewEthereumBlock:    false,
			BlockNumber:         1,
		},
		Events: nil,
	}

	TestMsgIndexSidecarErr = TestMsgIndexWithEvents{
		MsgIndex: &bridgetypes.MsgIndex{
			Authority:           TestGovernanceAddress,
			NumInjectedEventTxs: 0,
			NewEthereumBlock:    false,
			BlockNumber:         1,
		},
		Events: nil,
	}

	TestEmptySidecarResponse          = &sidecartypes.QueryBlockEventsResponse{Events: nil}
	TestSidecarResponse               = &sidecartypes.QueryBlockEventsResponse{Events: TestEvents}
	TestSidecarResponseReduced        = &sidecartypes.QueryBlockEventsResponse{Events: TestEventsReduced}
	TestSidecarResponseWithFourEvents = &sidecartypes.QueryBlockEventsResponse{
		Events: TestEventsWithFourEvents,
	}
	TestSidecarResponseInvalidDeposit   = &sidecartypes.QueryBlockEventsResponse{Events: TestEventsInvalidDeposit}
	TestSidecarResponseInvalidAuthorize = &sidecartypes.QueryBlockEventsResponse{Events: TestEventsInvalidAuthorize}
	TestSidecarResponseDepositOnly      = &sidecartypes.QueryBlockEventsResponse{Events: TestEventsDepositOnly}
	TestSidecarResponseAuthorizeOnly    = &sidecartypes.QueryBlockEventsResponse{Events: TestEventsAuthorizeOnly}
)

func init() {
	TestEvent1Msg = MustGetDepositMsgFromDepositEvent(TestCdc, TestGovernanceAddress, TestEvent1)
	TestEvent3Msg = MustGetDepositMsgFromDepositEvent(TestCdc, TestGovernanceAddress, TestEvent3)
	TestEvent4Msg = MustGetDepositMsgFromDepositEvent(TestCdc, TestGovernanceAddress, TestEvent4)
	TestEvent5Msg = MustGetDepositMsgFromDepositEvent(TestCdc, TestGovernanceAddress, TestEvent5)
	TestEvent6Msg = MustGetDepositMsgFromDepositEvent(TestCdc, TestGovernanceAddress, TestEvent6)
	TestEvent7Msg = MustGetDepositMsgFromDepositEvent(TestCdc, TestGovernanceAddress, TestEvent7)
	TestEvent8Msg = MustGetDepositMsgFromDepositEvent(TestCdc, TestGovernanceAddress, TestEvent8)
	TestEvent9Msg = MustGetDepositMsgFromDepositEvent(TestCdc, TestGovernanceAddress, TestEvent9)
	TestEvent10Msg = MustGetDepositMsgFromDepositEvent(TestCdc, TestGovernanceAddress, TestEvent10)
	TestEvent11Msg = MustGetDepositMsgFromDepositEvent(TestCdc, TestGovernanceAddress, TestEvent11)
	TestEvent12Msg = MustGetDepositMsgFromDepositEvent(TestCdc, TestGovernanceAddress, TestEvent12)
}
