package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/types"
)

func (k msgServer) PostBlob(
	goCtx context.Context,
	msg *types.MsgPostBlob,
) (*types.MsgPostBlobResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Retrieve sequencing parameters
	params := k.GetParams(ctx)
	maxBlobSizeBytes := params.MaxBlobSizeBytes
	gasPerBlobSize := params.GasPerBlobByte

	// Verify that the data satisfies a maximum transaction size (using MaxBlobSizeBytes)
	msgLength := uint64(len(msg.Data))
	if msgLength > maxBlobSizeBytes {
		return nil, errorsmod.Wrapf(
			types.ErrDataTooBig,
			"message size %d exceeds max blob size bytes %d",
			msgLength,
			maxBlobSizeBytes,
		)
	}

	// Check if a topic exists, if it doesn't create a new one.
	topic, found := k.GetTopic(ctx, msg.Topic)
	if !found {

		// Verify that the topic order from the message is 0
		if !math.ZeroInt().Equal(msg.Order) {
			return nil, errorsmod.Wrapf(
				types.ErrOrderNotMatching,
				"msg order %s doesn't match next topic order %s",
				msg.Order.String(),
				math.ZeroInt().String(),
			)
		}

		// If the topic is not found, create one from scratch
		topic = types.Topic{
			Id:    msg.Topic,
			Owner: msg.From,
			Order: math.ZeroInt(),
		}

	} else {

		// Verify if the topic owner matches that of msg from
		if topic.Owner != msg.From {
			return nil, errorsmod.Wrapf(
				types.ErrSenderNotOwner,
				"from address %s doesn't match topic owner %s",
				msg.From,
				topic.Owner,
			)
		}

		// Verify order is the next expected order for the topic
		nextTopicOrder := topic.Order.Add(math.OneInt())
		if !nextTopicOrder.Equal(msg.Order) {
			return nil, errorsmod.Wrapf(
				types.ErrOrderNotMatching,
				"msg order %s doesn't match next topic order %s",
				msg.Order.String(),
				nextTopicOrder.String(),
			)
		}

		// Increment the topic order
		topic.Order = nextTopicOrder
	}

	// Calculate the total gas to be consumed for the blob based on its size
	totalGas := msgLength * gasPerBlobSize
	ctx.GasMeter().ConsumeGas(totalGas, "PostBlob data size")

	// Set the topic
	k.SetTopic(ctx, topic)

	// Update the ethereum nonce
	nonce := k.bridgeKeeper.MustGetLastEthereumNonce(ctx).AddRaw(1)
	k.bridgeKeeper.SetLastEthereumNonce(ctx, nonce)

	return &types.MsgPostBlobResponse{
		Nonce: nonce,
		From:  msg.From,
		Topic: msg.Topic,
		Order: msg.Order,
		Data:  msg.Data,
	}, nil
}
