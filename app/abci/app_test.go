package abci_test

import (
	"fmt"
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/fuel-infrastructure/fuel-sequencer/app/abci"
	"github.com/fuel-infrastructure/fuel-sequencer/app/apptesting"
	sidecartestutil "github.com/fuel-infrastructure/fuel-sequencer/sidecar/testutil"
	"github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	"github.com/stretchr/testify/suite"
)

type AppTestSuite struct {
	apptesting.KeeperTestHelper
}

// GetTestProposalHandler simply returns a ProposalHandler with the specified sidecar client mock
func (s *AppTestSuite) GetTestProposalHandler(
	sidecarClientMock *sidecartestutil.MockAppSidecarClient,
) *abci.FuelSequencerProposalHandler {
	return abci.NewFuelSequencerProposalHandler(
		s.App.AppCodec(), s.App.Logger(), s.App.StakingKeeper, s.App, sidecarClientMock, s.App.BridgeKeeper,
	)
}

// CreateDummyTxs creates transactions for testing purposes. It uses the bank module's MsgSend to fill the transactions
func (s *AppTestSuite) CreateDummyTxs(amount uint64, gasLimit uint64) []sdk.Tx {
	var txs []sdk.Tx

	txBuilder := s.App.NewTxBuilder()

	for i := uint64(0); i < amount; i++ {
		tx := s.KeeperTestHelper.BuildTx(
			txBuilder,
			[]sdk.Msg{
				&banktypes.MsgSend{
					FromAddress: "test-from",
					ToAddress:   "test-to",
					Amount:      sdk.NewCoins(),
				},
			},
			signing.SignatureV2{PubKey: nil, Data: nil, Sequence: 0},
			fmt.Sprintf("%d", i),
			sdk.NewCoins(),
			gasLimit,
		)
		txs = append(txs, tx)
	}

	return txs
}

// CreateEncodedDummyTxs creates transactions for testing purposes and encodes them
func (s *AppTestSuite) CreateEncodedDummyTxs(amount uint64, gasLimit uint64) [][]byte {
	txEncoder := s.App.GetTxEncoder()
	var txs [][]byte

	rawTxs := s.CreateDummyTxs(amount, gasLimit)

	for _, tx := range rawTxs {
		encodedTx, err := txEncoder(tx)
		if err != nil {
			panic("could not encode dummy transaction")
		}
		txs = append(txs, encodedTx)
	}

	return txs
}

// EncodeMsgSupplyDeltaTx is a helper that encodes a sdk.Tx containing MsgSupplyDelta to bytes
func (s *AppTestSuite) EncodeMsgSupplyDeltaTx() []byte {

	// Construct Any from message.
	msgSupplyDeltaAny, err := codectypes.NewAnyWithValue(&bridgetypes.MsgSupplyDelta{
		Authority: s.App.BridgeKeeper.GetAuthority(),
	})
	if err != nil {
		panic("could not construct any from MsgSupplyDelta")
	}

	// Construct Tx Body with the message.
	txBodyBz, err := proto.Marshal(&txtypes.TxBody{
		Messages: []*codectypes.Any{msgSupplyDeltaAny},
	})
	if err != nil {
		panic("could not construct TxBody")
	}

	// Construct Auth Info with Fee to avoid nil pointer panics.
	authInfoBz, err := proto.Marshal(&txtypes.AuthInfo{
		Fee: &txtypes.Fee{
			GasLimit: 0,
		},
	})
	if err != nil {
		panic("could not construct AuthInfo")
	}

	// Construct final Tx.
	txRawBz, err := proto.Marshal(&txtypes.TxRaw{
		BodyBytes:     txBodyBz,
		AuthInfoBytes: authInfoBz,
		Signatures:    nil,
	})
	if err != nil {
		panic("could not construct TxRaw bytes")
	}

	return txRawBz
}

// EncodeEthEventsTx is a helper to encode EthEventsTx to bytes
func (s *AppTestSuite) EncodeEthEventsTx(tx *types.TestEthEventsTxWithEvents) (txs [][]byte) {
	ethEventsTxBz, err := tx.EthEventsTx.RawTxBytes()
	if err != nil {
		panic(err)
	}
	txs = append(txs, ethEventsTxBz)

	for _, event := range tx.Events {
		eventTx, err := event.RawTxBytes(s.App.AppCodec())
		if err != nil {
			panic("could not get raw tx bytes from event")
		}
		txs = append(txs, eventTx)
	}

	return
}

func (s *AppTestSuite) SetupTest() {
	s.Setup()
}

func TestAppTestSuite(t *testing.T) {
	suite.Run(t, new(AppTestSuite))
}
