package abci

import (
	"encoding/hex"
	"errors"
	"fmt"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	sidecarclient "github.com/fuel-infrastructure/fuel-sequencer/sidecar/client"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	bridgekeeper "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/keeper"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

type FuelSequencerProposalHandler struct {
	logger       log.Logger
	valStore     baseapp.ValidatorStore         // to get the current validators' pubkeys
	txSelector   TxSelector                     // a utility for checking whether a tx can be included in the proposal
	txVerifier   baseapp.ProposalTxVerifier     // a utility for transaction verification
	sidecar      sidecarclient.AppSidecarClient // a client to query the Sidecar service
	bridgeKeeper bridgekeeper.Keeper            // Bridge keeper

	// TODO: Any required objects need to go here
}

// NewFuelSequencerProposalHandler defines a custom FuelSequencer proposal handler object
func NewFuelSequencerProposalHandler(
	logger log.Logger,
	valStore baseapp.ValidatorStore,
	txVerifier baseapp.ProposalTxVerifier,
	sidecar sidecarclient.AppSidecarClient,
	bridgeKeeper bridgekeeper.Keeper,
) *FuelSequencerProposalHandler {
	// TODO: Add any required parameters
	return &FuelSequencerProposalHandler{
		logger:       logger,
		valStore:     valStore,
		txVerifier:   txVerifier,
		txSelector:   NewFuelSequencerTxSelector(),
		sidecar:      sidecar,
		bridgeKeeper: bridgeKeeper,
	}
}

// PrepareProposalHandler defines the logic that is executed by the block proposer when they are crafting a new block.
// This function needs to satisfy the following rules:
//
// 1. Can be non-deterministic
// 2. Transaction size cannot exceed RequestPrepareProposal.MaxTxBytes
// 3. Block gas cannot exceed BlockParams.MaxGas
//
// Notes:
//
// 1. Here we are assuming that the NoOp mempool is to be used, meaning that txs requested from CometBFT will simply
// be returned and not verified. It is recommended that the ProcessProposalHandler implements any verifications that
// PrepareProposalHandler implements, therefore, for the NoOp mempool, ProcessProposalHandler shouldn't implement any
// verification checks. In the case that a different mempool is implemented we must perform extra tx verification checks
// in PrepareProposalHandler and ProcessProposalHandler because we are no longer relying on CometBFT, and thus we must
// ensure that we are including valid transactions.
// Please refer to the following default handler implementation on Cosmos SDK main branch if in doubt:
// https://github.com/cosmos/cosmos-sdk/blob/a86a83f761383c1ea434925cddd199cd5a271303/baseapp/abci_utils.go#L199-L303
//
// 2. Any error raised by PrepareProposalHandler is caught in baseapp/abci.go, resulting into the req.Txs to be returned
// to CometBFT.
// Reference: https://github.com/cosmos/cosmos-sdk/blob/a248d05f70f4ad7b8ff7b521e3d23086867d07dc/baseapp/abci.go#L447-L451
func (h *FuelSequencerProposalHandler) PrepareProposalHandler() sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
		//blockHeight, found := h.bridgeKeeper.GetLastEthereumBlockSynced(ctx)
		//if !found {
		//	return nil, errors.New("could not get last Ethereum block synced from state")
		//}

		/**
		TODO: 1. Sidecar not catching up
		      2. Sequencer too fast
		      3. Connection issues with sidecar
		      All the above will cause a new round to start immediately as frendo requested
		*/

		response, err := h.sidecar.GetBlockEvents(ctx, &sidecartypes.QueryBlockEventsRequest{BlockNumber: "4"})
		h.logger.Info("try1", "ERR", err)
		h.logger.Info("try2", "RESPONSE", response)
		for _, event := range response.Events {
			h.logger.Info("try3", "event type", event.EventType)
			var eventData sidecartypes.AuthorizeEvent
			err = eventData.Unmarshal(event.Data)
			h.logger.Info("try4", "ERR", err)
			h.logger.Info("try5", "event data", eventData)
			h.logger.Info("try6", "from hex", eventData.From)
			h.logger.Info("try7", "from", common.HexToAddress(eventData.From))
			h.logger.Info("try8", "msg", hex.EncodeToString(eventData.Message))
		}

		response, err = h.sidecar.GetBlockEvents(ctx, &sidecartypes.QueryBlockEventsRequest{BlockNumber: "3"})
		h.logger.Info("try9", "ERR", err)
		h.logger.Info("try10", "RESPONSE", response)
		for _, event := range response.Events {
			h.logger.Info("try11", "event type", event.EventType)
			var eventData sidecartypes.SendToSequencerEvent
			err = eventData.Unmarshal(event.Data)
			h.logger.Info("try12", "ERR", err)
			h.logger.Info("try13", "event data", eventData)
			h.logger.Info("try14", "from hex", eventData.From)
			h.logger.Info("try15", "from", common.HexToAddress(eventData.From))
			h.logger.Info("try16", "to", eventData.To)
			duration, success := sdkmath.NewIntFromString(eventData.Duration)
			h.logger.Info("try17", "success", success)
			h.logger.Info("try18", "duration", duration)
			amount, success := sdkmath.NewIntFromString(eventData.Amount)
			h.logger.Info("try19", "success", success)
			h.logger.Info("try20", "amount", amount)
		}

		// TODO: This should be adapted as per application requirements
		ethEventsTx, err := h.generateEthEventsTx()
		if err != nil {
			return nil, fmt.Errorf("failed to generate eth events tx: %w", err)
		}
		ethEventsTxBz, err := ethEventsTx.Marshal()
		if err != nil {
			return nil, fmt.Errorf("failed to encode injected eth events tx: %w", err)
		}
		// Current implementation assumes txs that cannot be decoded into sdk.Tx are appended at entry 0.
		req.Txs = append([][]byte{ethEventsTxBz}, req.Txs...)

		// TODO: Define custom logic here

		var maxBlockGas uint64
		if b := ctx.ConsensusParams().Block; b != nil {
			maxBlockGas = uint64(b.MaxGas)
		}

		// This is needed to clear variables tracked by the TxSelector
		defer h.txSelector.Clear()

		// Since we are assuming a NoOp mempool we simply return the transactions requested from CometBFT, which, by
		// default, should be in FIFO order. Note, we still need to ensure the transactions returned respect
		// req.MaxTxBytes and blockParams.MaxGas
		for index, txBz := range req.Txs {

			// Eth events tx will not satisfy sdk.Tx. Therefore, we cannot handle it the same way as we handle other
			// transactions. Note: Here we are assuming that the eth events tx is injected at index zero and that we
			// will always have such a tx.
			if index == 0 {
				success := h.txSelector.SelectNonSDKTxForProposal(ctx, uint64(req.MaxTxBytes), txBz)
				if !success {
					return nil, errors.New("failed to add eth events transaction to block proposal")
				}
				continue
			}

			tx, err := h.txVerifier.TxDecode(txBz)
			if err != nil {
				return nil, err
			}

			stop := h.txSelector.SelectTxForProposal(ctx, uint64(req.MaxTxBytes), maxBlockGas, tx, txBz)
			if stop {
				break
			}
		}

		h.logger.Debug("prepared proposal", "txs", len(h.txSelector.SelectedTxs(ctx)))

		return &abci.ResponsePrepareProposal{Txs: h.txSelector.SelectedTxs(ctx)}, nil
	}
}

// ProcessProposalHandler defines the logic which confirms the validity of a block proposal. This logic is executed
// by all validators, and it needs to satisfy the following:
//
// 1. Must be deterministic
// 2. Block gas cannot exceed BlockParams.MaxGas
//
// Notes:
//
// 1. Here we are assuming that the NoOp mempool is to be used, meaning that PrepareProposal may include some txs
// that might fail verification. The logic implemented here attempts to perform the exact verifications performed by the
// PrepareProposalHandler as suggested by the Cosmos SDK documentation. Unfortunately, we cannot verify whether
// transaction sizes exceed the limit because MaxTxBytes is not part of abci.RequestProcessProposal. That being said, it
// seems that CometBFT validates the Tx sizes returned by the application https://github.com/cometbft/cometbft/blob/719b64156aaa3cb89add29d053439060f8e420dd/spec/consensus/creating-proposal.md?plain=1#L41
// It is also important to highlight that the default implementation of ProcessProposalHandler doesn't perform any
// verifications for NoOp mempools, however, we have included a block gas max limit check to follow recommendations
// stated in the documentation.
// Please refer to the following default handler implementation on Cosmos SDK main branch if in doubt:
// https://github.com/cosmos/cosmos-sdk/blob/a86a83f761383c1ea434925cddd199cd5a271303/baseapp/abci_utils.go#L305-L352
//
// 2. Any error raised by ProcessProposalHandler is caught in baseapp/abci.go, resulting into the application to reject
// the block proposal.
// Reference: https://github.com/cosmos/cosmos-sdk/blob/a248d05f70f4ad7b8ff7b521e3d23086867d07dc/baseapp/abci.go#L541-L545
func (h *FuelSequencerProposalHandler) ProcessProposalHandler() sdk.ProcessProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
		// TODO: This should be adapted as per application requirements
		// First expect that the first transaction is always the EthEventsTx
		var injectedEthEventsTx bridgetypes.EthEventsTx
		if err := injectedEthEventsTx.Unmarshal(req.Txs[0]); err != nil {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
		}

		// TODO: Perform any other verifications on injected txs

		// TODO: This should be adapted as per application requirements
		ethEventsTx, err := h.generateEthEventsTx()
		if err != nil {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
		}
		if !injectedEthEventsTx.Equal(&ethEventsTx) {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
		}

		// TODO: Define custom logic here

		var totalTxGas uint64

		var maxBlockGas int64
		if b := ctx.ConsensusParams().Block; b != nil {
			maxBlockGas = b.MaxGas
		}

		for index, txBytes := range req.Txs {
			// Eth events transactions typically don't implement sdk.Tx, therefore, they can be skipped since no gas
			// is consumed. Note: Here we are assuming that eth events txs are always injected at index zero.
			if index == 0 {
				continue
			}

			// There is something wrong with the Tx if it cannot be decoded. Thus reject the block proposal.
			tx, err := h.txVerifier.TxDecode(txBytes)
			if err != nil {
				return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
			}

			// Confirm that the block's max gas limit is not exceeded
			if maxBlockGas > 0 {
				gasTx, ok := tx.(baseapp.GasTx)
				if ok {
					totalTxGas += gasTx.GetGas()
				}

				if totalTxGas > uint64(maxBlockGas) {
					return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
				}
			}
		}

		h.logger.Debug("processing proposal", "height", req.Height, "num_txs", len(req.Txs))

		return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
	}
}

// TODO: This function should be replaced with our application logic
func (h *FuelSequencerProposalHandler) generateEthEventsTx() (bridgetypes.EthEventsTx, error) {
	// TODO: Perform any custom logic

	return bridgetypes.EthEventsTx{
		EventsData: []string{"Event 1", "Event 2", "Event 3"},
	}, nil
}

// PreBlocker contains logic that should run before any FinalizeBlock logic. FinalizeBlock ignores any byte slices not
// implementing sdk.Tx. As a result, any important results originating from PrepareProposal or ProcessProposal not
// implementing sdk.Tx need to be made available to the modules in storage at PreBlocker stage.
func (h *FuelSequencerProposalHandler) PreBlocker(
	_ sdk.Context, req *abci.RequestFinalizeBlock,
) (*sdk.ResponsePreBlock, error) {
	// TODO: This should be adapted as per application requirements
	// This check is done for completeness’s sake as we should not expect to run into this scenario
	if len(req.Txs) == 0 {
		return nil, fmt.Errorf("expected eth events transaction to be injected")
	}

	// TODO: Check if certain transactions are expected at this stage ex MsgSupplyDelta at specific epochs

	// TODO: This was done for demonstration purposes and should be adapted as per application requirements.
	var injectedEthEventsTx bridgetypes.EthEventsTx
	if err := injectedEthEventsTx.Unmarshal(req.Txs[0]); err != nil {
		return nil, fmt.Errorf("failed to decode injected eth events tx: %w", err)
	}

	// TODO: Custom logic like storing "special" transactions in state

	h.logger.Debug("finished executing pre-block hook")

	return &sdk.ResponsePreBlock{}, nil
}
