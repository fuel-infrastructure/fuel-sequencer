package abci

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/log"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type FuelSequencerVoteExtHandler struct {
	logger log.Logger

	// Any required objects need to go here
}

func NewFuelSequencerVoteExtHandler(logger log.Logger) *FuelSequencerVoteExtHandler {
	return &FuelSequencerVoteExtHandler{
		logger: logger,
	}
}

// TODO: This needs to be replaced by our application logic
func (h *FuelSequencerVoteExtHandler) getVoteExtensionData(height int64) CustomData {
	msgSend := "CooBChwvY29zbW9zLmJhbmsudjFiZXRhMS5Nc2dTZW5kEmoKLWNvc21vczF2dGZ6cms2ZjRtNmt4dDZlaHlxdDlqNXN1NWh2Y3o1cXhhcGtzZRItY29zbW9zMWczdjNhMmM1cWs4Znc4enFqbXZjbnNuMDJrMnVscmtnaDU2NDZwGgoKBXRva2VuEgEx"
	return CustomData{
		Data: []byte(fmt.Sprintf("HELLO (%d)", height)),
		Msgs: msgSend,
	}
}

// ExtendVoteHandler implements the business logic that allows an application to extend a pre-commit vote with arbitrary
// data. The ExtendVoteHandler must obey these rules:
//
// 1. It can be non-deterministic.
// 2. VoteExtension in *abci.ResponseExtendVote must never be nil.
// 3. VoteExtension bytes can be empty.
//
// Note: It is important to keep vote extension sizes minimal to avoid delays in block production. The Cosmos SDK
// documentation advises against using JSON for encoding vote extensions, favoring more compact formats to reduce output
// size.
func (h *FuelSequencerVoteExtHandler) ExtendVoteHandler() sdk.ExtendVoteHandler {
	return func(ctx sdk.Context, req *abci.RequestExtendVote) (*abci.ResponseExtendVote, error) {

		// TODO: This should be removed as it was implemented for demonstration purposes
		voteExt := CustomOracleVoteExtension{
			Height: req.Height,
			Data:   h.getVoteExtensionData(req.Height),
		}
		bz, err := json.Marshal(voteExt)
		if err != nil {
			return &abci.ResponseExtendVote{VoteExtension: []byte{}}, fmt.Errorf(
				"failed to marshal vote extension: %w", err)
		}

		// TODO: Define custom logic here

		h.logger.Info(
			"extending vote with oracle prices",
			"req_height", req.Height,
		)

		return &abci.ResponseExtendVote{VoteExtension: bz}, nil
	}
}

// TODO: Add inline comment
func (h *FuelSequencerVoteExtHandler) VerifyVoteExtensionHandler() sdk.VerifyVoteExtensionHandler {
	// TODO: Verify size in here'
	// TODO: todo for custom logic
	// TODO: Add boiler plate logic
	return func(ctx sdk.Context, req *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error) {
		return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_ACCEPT}, nil
	}
}
