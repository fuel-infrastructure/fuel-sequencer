package keeper

import (
	"context"
	"crypto/ed25519"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/fuel-infrastructure/fuel-sequencer/x/blob/types"
)

// SubmitDAAttestation handles MsgSubmitDAAttestation — verifies Ed25519 attestation
// signatures against the consensus validator set, weighted by voting power (55% threshold).
func (k msgServer) SubmitDAAttestation(
	goCtx context.Context, msg *types.MsgSubmitDAAttestation,
) (*types.MsgSubmitDAAttestationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if k.stakingKeeper == nil {
		return nil, types.ErrInvalidAttestation.Wrap("staking keeper not configured")
	}

	// 1. Get active validator set
	validators, err := k.stakingKeeper.GetLastValidators(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get validator set: %w", err)
	}

	// 2. Build lookup: consensus address (hex) → {ed25519 pubkey bytes, voting power}
	type valInfo struct {
		pubKeyBytes []byte
		power       int64
	}
	valLookup := make(map[string]valInfo, len(validators))
	var totalPower int64

	for _, v := range validators {
		consPubKey, err := v.ConsPubKey()
		if err != nil {
			continue
		}
		// CometBFT consensus address is the first 20 bytes of the pubkey hash
		consAddr := sdk.ConsAddress(consPubKey.Address())
		consAddrHex := hex.EncodeToString(consAddr)

		power := v.GetConsensusPower(sdk.DefaultPowerReduction)
		totalPower += power

		valLookup[consAddrHex] = valInfo{
			pubKeyBytes: consPubKey.Bytes(),
			power:       power,
		}
	}

	if totalPower == 0 {
		return nil, types.ErrInvalidAttestation.Wrap("total voting power is zero")
	}

	// 3. Verify each attestation signature and track unique valid attestors
	validAttestors := make(map[string]int64) // consAddrHex → power (deduplicates)

	for _, att := range msg.Attestations {
		vi, found := valLookup[att.ValidatorAddress]
		if !found {
			ctx.Logger().Debug("unknown validator in DA attestation",
				"validator_address", att.ValidatorAddress)
			continue
		}

		// Already counted this validator
		if _, already := validAttestors[att.ValidatorAddress]; already {
			continue
		}

		// Construct canonical 68-byte message: blob_key(32) || chunk_index(4 BE) || chunk_hash(32)
		attMsg := make([]byte, 68)
		copy(attMsg[0:32], msg.BlobKey)
		binary.BigEndian.PutUint32(attMsg[32:36], att.ChunkIndex)
		copy(attMsg[36:68], att.ChunkHash)

		// Ed25519 verify
		pubKey := ed25519.PublicKey(vi.pubKeyBytes)
		if !ed25519.Verify(pubKey, attMsg, att.Signature) {
			ctx.Logger().Debug("invalid Ed25519 signature in DA attestation",
				"validator_address", att.ValidatorAddress,
				"chunk_index", att.ChunkIndex)
			continue
		}

		validAttestors[att.ValidatorAddress] = vi.power
	}

	// 4. Sum voting power of unique valid attestors
	var attestedPower int64
	for _, power := range validAttestors {
		attestedPower += power
	}

	// 5. Check >= 55% of total power
	// Use integer math: attestedPower * 100 >= totalPower * 55
	confirmed := attestedPower*100 >= totalPower*55

	totalPowerStr := fmt.Sprintf("%d", totalPower)
	attestedPowerStr := fmt.Sprintf("%d", attestedPower)

	// 6. Emit event
	eventType := types.EventTypeDAAttestationConfirmed
	if !confirmed {
		eventType = types.EventTypeDAAttestationRejected
	}

	blobKeyHex := hex.EncodeToString(msg.BlobKey)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			eventType,
			sdk.NewAttribute(types.AttributeKeyBlobHash, blobKeyHex),
			sdk.NewAttribute(types.AttributeKeyBlobSize, fmt.Sprintf("%d", msg.BlobSize)),
			sdk.NewAttribute(types.AttributeKeyTotalPower, totalPowerStr),
			sdk.NewAttribute(types.AttributeKeyAttestedPower, attestedPowerStr),
			sdk.NewAttribute(types.AttributeKeyConfirmed, fmt.Sprintf("%t", confirmed)),
		),
	)

	// Guard against int overflow when computing percent
	var pct float64
	if totalPower > 0 {
		pct = float64(attestedPower) / float64(totalPower) * 100
	}
	if math.IsInf(pct, 0) || math.IsNaN(pct) {
		pct = 0
	}

	ctx.Logger().Info("DA attestation processed",
		"blob_key", blobKeyHex,
		"confirmed", confirmed,
		"attested_power", attestedPowerStr,
		"total_power", totalPowerStr,
		"valid_attestors", len(validAttestors),
		"total_attestations", len(msg.Attestations),
		"pct", fmt.Sprintf("%.1f%%", pct),
	)

	return &types.MsgSubmitDAAttestationResponse{
		Confirmed:     confirmed,
		TotalPower:    totalPowerStr,
		AttestedPower: attestedPowerStr,
	}, nil
}
