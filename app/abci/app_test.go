package abci_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/fuel-infrastructure/fuel-sequencer/app/abci"
	"github.com/fuel-infrastructure/fuel-sequencer/app/apptesting"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	sidecartestutil "github.com/fuel-infrastructure/fuel-sequencer/sidecar/testutil"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	"github.com/stretchr/testify/suite"
)

type AppTestSuite struct {
	apptesting.KeeperTestHelper
}

type TestQueryBlockEventsRet struct {
	Response *sidecartypes.QueryBlockEventsResponse
	Error    error
}

// GetTestProposalHandler simply returns a ProposalHandler with the specified sidecar client mock
func (s *AppTestSuite) GetTestProposalHandler(
	sidecarClientMock *sidecartestutil.MockAppSidecarClient,
) *abci.FuelSequencerProposalHandler {
	return abci.NewFuelSequencerProposalHandler(
		s.App.Logger(), s.App.StakingKeeper, s.App, sidecarClientMock, s.App.BridgeKeeper,
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

// EncodeEthEventsTx is a helper to encode EthEventsTx to bytes
func (s *AppTestSuite) EncodeEthEventsTx(tx *bridgetypes.EthEventsTx) []byte {
	ethEventsTxBz, err := tx.Marshal()
	if err != nil {
		panic("could not encode eth events tx")
	}

	return ethEventsTxBz
}

func (s *AppTestSuite) SetupTest() {
	s.Setup()
}

func TestAppTestSuite(t *testing.T) {
	suite.Run(t, new(AppTestSuite))
}
