package utils_test

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/gogoproto/proto"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func TestProtoSerialization(t *testing.T) {
	// Construct the MsgSend with your specified addresses
	amt, _ := sdkmath.NewIntFromString("10")
	msgsend := &banktypes.MsgSend{
		FromAddress: "fuelsequencer13hfdkxj5aeqzsll569mqreedkafp34ngcsjkqjmpq6prtgv80kcq83gttw",
		ToAddress:   "fuelsequencer163rsv65t4893t2rz5rmda9sly7lgdlq2jgr36m",
		Amount: []sdk.Coin{
			{Denom: "ufuel", Amount: amt}, // Example amount, adjust as needed
		},
	}
	anymsgsend, _ := codectypes.NewAnyWithValue(msgsend)

	// Serialize the message to bytes
	data, _ := proto.Marshal(&bridgetypes.AuthorizeTx{Messages: []*codectypes.Any{anymsgsend}})

	// Convert the serialized bytes to a hex string
	hexData := fmt.Sprintf("0x%x", data)
	fmt.Println(fmt.Sprintf("Serialized Hex Data: %s", hexData))
}
