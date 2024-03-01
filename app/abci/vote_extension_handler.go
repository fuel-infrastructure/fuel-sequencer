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

// VerifyVoteExtensionHandler implements some checks ensuring that vote extensions submitted in the pre-commits satisfy
// important criteria. The VerifyVoteExtensionHandler must be deterministic and should thoroughly confirm that the vote
// extension is valid. Something important to keep in mind is that validators do not verify the vote extensions of all
// other validators because this depends highly on how the vote extensions get propagated. As a result,
// this functionality should not replace any verification done in abci.ProcessProposal or abci.PrepareProposal, but
// should complement it.
func (h *FuelSequencerVoteExtHandler) VerifyVoteExtensionHandler() sdk.VerifyVoteExtensionHandler {
	return func(ctx sdk.Context, req *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error) {

		// TODO: The following check must be done in production but we need to replace with our application logic.
		// Unmarshal the vote extension to confirm that it was formatted correctly
		var voteExt CustomOracleVoteExtension
		err := json.Unmarshal(req.VoteExtension, &voteExt)
		if err != nil {
			return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT},
				fmt.Errorf("failed to unmarshal vote extension: %w", err)
		}

		// TODO: The following check must be done in production but we need to replace with our application logic.
		// Confirm that the vote extension was submitted at the right height
		if voteExt.Height != req.Height {
			return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT},
				fmt.Errorf(
					"vote extension height does not match request height; expected: %d, got: %d", req.Height,
					voteExt.Height,
				)
		}

		// TODO: The following check must be done in production but we need to replace with our application logic.
		// Validate the vote extension data
		if err := h.verifyVoteExtensionData(ctx, voteExt.Data); err != nil {
			return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT},
				fmt.Errorf("failed to verify oracle prices from validator %X: %w", req.ValidatorAddress, err)
		}

		// Confirm that the vote extension does not exceed the maximum size. This prevents performance degradation
		if len(req.VoteExtension) > MaxVESize {
			return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT},
				fmt.Errorf("vote extension is too large %d: limit %d", len(req.VoteExtension), MaxVESize)
		}

		// TODO: Add custom logic here

		h.logger.Info(
			"validated vote extension",
			"height", req.Height,
			"size (bytes)", len(req.VoteExtension),
		)

		return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_ACCEPT}, nil
	}
}

// verifyVoteExtensionData implements specific checks to ensure the integrity and validity of the data within a vote
// extension.
func (h *FuelSequencerVoteExtHandler) verifyVoteExtensionData(ctx sdk.Context, data CustomData) error {
	// TODO: Custom application logic that verifies the vote extension data should be added here
	return nil
}
