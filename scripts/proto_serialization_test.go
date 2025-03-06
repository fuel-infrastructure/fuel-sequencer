package scripts_test

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cometbft/cometbft/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/ethereum/go-ethereum/accounts/abi"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	testutiltypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/fuel-infrastructure/fuel-sequencer/utils"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func TestProtoSerialization_MsgSend(t *testing.T) {

	// Construct the MsgSend with your specified addresses
	amt, _ := sdkmath.NewIntFromString("1")
	msgSend := &banktypes.MsgSend{
		FromAddress: "0x2B4ce813f1e814004c7B806bC31B4Fe0650C6FA8",
		ToAddress:   "0x2B4ce813f1e814004c7B806bC31B4Fe0650C6FA8",
		Amount: []sdk.Coin{
			{Denom: "utest", Amount: amt}, // Example amount, adjust as needed
		},
	}
	anyMsgSend, _ := codectypes.NewAnyWithValue(msgSend)

	// Serialize the message to bytes
	messages := []*codectypes.Any{
		anyMsgSend,
	}
	data, _ := proto.Marshal(&bridgetypes.AuthorizeTx{Messages: messages})
	fmt.Printf("Transaction size (bytes): %d\n", utils.TxSize(data))

	// Convert the serialized bytes to a hex string
	hexData := fmt.Sprintf("0x%s", hex.EncodeToString(data))
	fmt.Printf("Serialized Hex Data: %s\n", hexData)
}

func TestProtoSerialization_VarietyOfMessages(t *testing.T) {

	// Construct the MsgSend with your specified addresses
	amt, _ := sdkmath.NewIntFromString("100")
	msgDelegate := &stakingtypes.MsgDelegate{
		DelegatorAddress: "0x2B4ce813f1e814004c7B806bC31B4Fe0650C6FA8",
		ValidatorAddress: "fuelsequencervaloper1cv0rl38sckgwyrkdd5vanyzf6v8clf809f74ca",
		Amount: sdk.Coin{
			Denom: "utest", Amount: amt, // Example amount, adjust as needed
		},
	}
	msgBeginRedelegate := &stakingtypes.MsgBeginRedelegate{
		DelegatorAddress:    "0x2B4ce813f1e814004c7B806bC31B4Fe0650C6FA8",
		ValidatorSrcAddress: "fuelsequencervaloper1cv0rl38sckgwyrkdd5vanyzf6v8clf809f74ca",
		ValidatorDstAddress: "fuelsequencervaloper1ddjv8z30raavjc8ku6n6mqlm9rjhezs27h8g6f",
		Amount: sdk.Coin{
			Denom: "utest", Amount: amt.QuoRaw(2), // Example amount, adjust as needed
		},
	}
	msgUndelegate := &stakingtypes.MsgUndelegate{
		DelegatorAddress: "0x2B4ce813f1e814004c7B806bC31B4Fe0650C6FA8",
		ValidatorAddress: "fuelsequencervaloper1cv0rl38sckgwyrkdd5vanyzf6v8clf809f74ca",
		Amount: sdk.Coin{
			Denom: "utest", Amount: amt.QuoRaw(2), // Example amount, adjust as needed
		},
	}
	msgWithdrawToEthereum := &bridgetypes.MsgWithdrawToEthereum{
		From: "0x2B4ce813f1e814004c7B806bC31B4Fe0650C6FA8",
		To:   "0x2B4ce813f1e814004c7B806bC31B4Fe0650C6FA8",
		Amount: sdk.Coin{
			Denom: "utest", Amount: amt, // Example amount, adjust as needed
		},
	}
	msgSend := &banktypes.MsgSend{
		FromAddress: "0x2B4ce813f1e814004c7B806bC31B4Fe0650C6FA8",
		ToAddress:   "0x2B4ce813f1e814004c7B806bC31B4Fe0650C6FA8",
		Amount: []sdk.Coin{
			{Denom: "utest", Amount: amt}, // Example amount, adjust as needed
		},
	}
	anyMsgDelegate, _ := codectypes.NewAnyWithValue(msgDelegate)
	anyMsgBeginRedelegate, _ := codectypes.NewAnyWithValue(msgBeginRedelegate)
	anyMsgUndelegate, _ := codectypes.NewAnyWithValue(msgUndelegate)
	anyMsgWithdrawToEthereum, _ := codectypes.NewAnyWithValue(msgWithdrawToEthereum)
	anyMsgSend, _ := codectypes.NewAnyWithValue(msgSend)

	// Serialize the message to bytes
	messages := []*codectypes.Any{
		anyMsgDelegate,
		anyMsgBeginRedelegate,
		anyMsgUndelegate,
		anyMsgWithdrawToEthereum,
		anyMsgSend,
	}
	data, _ := proto.Marshal(&bridgetypes.AuthorizeTx{Messages: messages})
	fmt.Printf("Transaction size (bytes): %d\n", utils.TxSize(data))

	// Convert the serialized bytes to a hex string
	hexData := fmt.Sprintf("0x%s", hex.EncodeToString(data))
	fmt.Printf("Serialized Hex Data: %s\n", hexData)
}

func TestDecodeDepositEventFromSidecar(t *testing.T) {

	dataBase64 := "CioweDAwNkExNzU2YWI1NzFhOWM5NjFkMjk2NTU3YmY2MGM1MGQ0OGE1MDASKjB4MDA2QTE3NTZhYjU3MWE5Yzk2MWQyOTY1NTdiZjYwYzUwZDQ4YTUwMBoTNTAwMDAwMDAwMDAwMDAwMDAwMCIDMzAw"
	dataBz, err := base64.StdEncoding.DecodeString(dataBase64)
	if err != nil {
		panic(err)
	}

	var eventData sidecartypes.DepositEvent
	err = eventData.Unmarshal(dataBz)
	if err != nil {
		panic(err)
	}

	fmt.Println(fmt.Sprintf(""+
		"Depositor: %s\n"+
		"Recipient: %s\n"+
		"Amount: %s\n"+
		"Lockup: %s\n",
		eventData.Depositor, eventData.Recipient, eventData.Amount, eventData.Lockup,
	))
}

func TestEncodeDepositEvent(t *testing.T) {

	eventData := sidecartypes.DepositEvent{
		Depositor: "0x006A1756ab571a9c961d296557bf60c50d48a500",
		Recipient: "0x006A1756ab571a9c961d296557bf60c50d48a500",
		Amount:    "5000000000",
		Lockup:    "300",
	}

	dataBz, err := eventData.Marshal()
	if err != nil {
		panic(err)
	}
	dataBase64 := base64.StdEncoding.EncodeToString(dataBz)

	fmt.Println(dataBase64)
}

func TestDecodeAuthorizeEventData(t *testing.T) {

	// Authorize event data
	hexData := "0a85010a1c2f636f736d6f732e62616e6b2e763162657461312e4d736753656e6412650a2a307866333966643665353161616438386636663463653661623838323732373963666666623932323636122a3078643434373036366138626139636231356138363261306636646539363166323762653836666330611a0b0a05756675656c12023130"
	hexDataBz, err := hex.DecodeString(hexData)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Transaction size (bytes): %d\n", utils.TxSize(hexDataBz))

	var authorizeTx bridgetypes.AuthorizeTx
	err = authorizeTx.Unmarshal(hexDataBz)
	if err != nil {
		panic(err)
	}

	// Encode as valid transaction
	rawTxBz, err := utils.ValidRawTxBytesFromAnyMsgs(authorizeTx.Messages, 0)
	if err != nil {
		panic(err)
	}

	// Decode again to extract SDK messages
	tx, err := authtx.DefaultTxDecoder(testutiltypes.TestCdc)(rawTxBz)
	if err != nil {
		panic(err)
	}

	for i, msg := range tx.GetMsgs() {
		fmt.Printf("MSG %d (%s): %s\n", i, sdk.MsgTypeURL(msg), msg)
	}
}

func TestDecodeAuthorizeEventDataFromEthereum(t *testing.T) {

	// Authorize event data (e.g. https://sepolia.etherscan.io/tx/0xcf8c887f060d6f6c9b5d80098bac6db2249749bcc25fd96c0e6133b9f84e03be#eventlog)
	hexData := "000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000970a94010a232f636f736d6f732e7374616b696e672e763162657461312e4d736744656c6567617465126d0a2a307837373935366536626338666536396132373162353263633435376162313139306465346133363062122a3078363264323231646234396165663536333266353962393030623263613930653532656363306138301a130a0474657374120b3130303030303030303030000000000000000000"
	hexDataBz, err := hex.DecodeString(hexData)
	if err != nil {
		panic(err)
	}

	// Contract ABI
	var contractAbi abi.ABI
	err = contractAbi.UnmarshalJSON([]byte(sidecartypes.SequencerProxyContractABI))
	if err != nil {
		panic(err)
	}

	// Parse the event data according to the ABI
	var ethEvent sidecartypes.EthAuthorizeEvent
	err = contractAbi.UnpackIntoInterface(&ethEvent, sidecartypes.AuthorizeEventName, hexDataBz)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Transaction size (bytes): %d\n", utils.TxSize(ethEvent.Data))

	var authorizeTx bridgetypes.AuthorizeTx
	err = authorizeTx.Unmarshal(ethEvent.Data)
	if err != nil {
		panic(err)
	}

	// Encode as valid transaction
	rawTxBz, err := utils.ValidRawTxBytesFromAnyMsgs(authorizeTx.Messages, 0)
	if err != nil {
		panic(err)
	}

	// Decode again to extract SDK messages
	tx, err := authtx.DefaultTxDecoder(testutiltypes.TestCdc)(rawTxBz)
	if err != nil {
		panic(err)
	}

	for i, msg := range tx.GetMsgs() {
		fmt.Printf("MSG %d (%s): %s\n", i, sdk.MsgTypeURL(msg), msg)
	}
}

func TestDecodeTx_Base64(t *testing.T) {

	dataBase64 := "Cl8KXQohL2Z1ZWxzZXF1ZW5jZXIuYnJpZGdlLnYxLk1zZ0luZGV4EjgKNGZ1ZWxzZXF1ZW5jZXIxMGQwN3kyNjVnbW11dnQ0ejB3OWF3ODgwam5zcjcwMGpkamZ2azMgARICEgA="
	dataBz, err := base64.StdEncoding.DecodeString(dataBase64)
	if err != nil {
		panic(err)
	}
	fmt.Printf("SIZE: %d\n", utils.TxSize(dataBz))
	fmt.Printf("HASH: %X\n", types.Tx(dataBz).Hash())

	tx, err := authtx.DefaultTxDecoder(testutiltypes.TestCdc)(dataBz)
	if err != nil {
		panic(err)
	}

	for i, msg := range tx.GetMsgs() {
		fmt.Printf("MSG %d (%s): %s\n", i, sdk.MsgTypeURL(msg), msg)
	}
}

func TestDecodeTx_Hex(t *testing.T) {

	dataHex := "0a620a600a212f6675656c73657175656e6365722e6272696467652e76312e4d7367496e646578123b0a346675656c73657175656e636572313064303779323635676d6d757674347a30773961773838306a6e73723730306a646a66766b3320c4b6850312021200"
	dataBz, err := hex.DecodeString(dataHex)
	if err != nil {
		panic(err)
	}
	fmt.Printf("SIZE: %d\n", utils.TxSize(dataBz))
	fmt.Printf("HASH: %X\n", types.Tx(dataBz).Hash())

	tx, err := authtx.DefaultTxDecoder(testutiltypes.TestCdc)(dataBz)
	if err != nil {
		panic(err)
	}

	for i, msg := range tx.GetMsgs() {
		fmt.Printf("MSG %d (%s): %s\n", i, sdk.MsgTypeURL(msg), msg)
	}
}
