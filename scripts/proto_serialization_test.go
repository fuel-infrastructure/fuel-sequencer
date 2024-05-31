package scripts_test

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/gogoproto/proto"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	testutiltypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func TestProtoSerialization(t *testing.T) {
	// Construct the MsgSend with your specified addresses
	amt, _ := sdkmath.NewIntFromString("10")
	msgsend := &banktypes.MsgSend{
		FromAddress: "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
		ToAddress:   "0xd447066a8ba9cb15a862a0f6de961f27be86fc0a",
		Amount: []sdk.Coin{
			{Denom: "ufuel", Amount: amt}, // Example amount, adjust as needed
		},
	}
	anymsgsend, _ := codectypes.NewAnyWithValue(msgsend)

	// Serialize the message to bytes
	data, _ := proto.Marshal(&bridgetypes.AuthorizeTx{Messages: []*codectypes.Any{anymsgsend}})

	// Convert the serialized bytes to a hex string
	hexData := fmt.Sprintf("0x%s", hex.EncodeToString(data))
	fmt.Printf("Serialized Hex Data: %s\n", hexData)
}

func TestDecodeDepositEvent(t *testing.T) {

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
		Amount:    "5000000000000000000",
		Lockup:    "300",
	}

	dataBz, err := eventData.Marshal()
	if err != nil {
		panic(err)
	}
	dataBase64 := base64.StdEncoding.EncodeToString(dataBz)

	fmt.Println(dataBase64)
}

func TestDecodeTx(t *testing.T) {

	dataBase64 := "Cl8KXQohL2Z1ZWxzZXF1ZW5jZXIuYnJpZGdlLnYxLk1zZ0luZGV4EjgKNGZ1ZWxzZXF1ZW5jZXIxMGQwN3kyNjVnbW11dnQ0ejB3OWF3ODgwam5zcjcwMGpkamZ2azMgARICEgA="
	dataBz, err := base64.StdEncoding.DecodeString(dataBase64)
	if err != nil {
		panic(err)
	}

	tx, err := authtx.DefaultTxDecoder(testutiltypes.TestCdc)(dataBz)
	if err != nil {
		panic(err)
	}

	for i, msg := range tx.GetMsgs() {
		fmt.Printf("MSG %d: %s\n", i, msg)
	}
}
