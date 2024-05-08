package keeper

import (
	"context"
	"strings"

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
	maxBlobSizeBytes := k.GetParams(ctx).MaxBlobSizeBytes

	// Verify that the data satisfies a maximum transaction size (using MaxBlobSizeBytes)
	msgLength := uint64(len(msg.Data))
	if msgLength > maxBlobSizeBytes {
		return nil, types.ErrDataTooBig.Wrapf(
			"message size %d exceeds max blob size bytes %d",
			msgLength, maxBlobSizeBytes,
		)
	}

	// Check if a topic exists, if it doesn't create a new one.
	topic, found := k.GetTopic(ctx, msg.Topic)
	if !found {

		// Verify that the topic order from the message is 0
		if !msg.Order.IsZero() {
			return nil, types.ErrOrderNotMatching.Wrapf(
				"msg order %s expected to be 0 for new topics",
				msg.Order.String(),
			)
		}

		// If the topic is not found, create one from scratch
		topic = types.Topic{
			Id:    msg.Topic,
			Owner: msg.From,
			Order: math.ZeroInt(),
		}
		if err := topic.ValidateBasic(); err != nil {
			return nil, types.ErrTopicFailedValidate.Wrapf("topic failed to validate basic: %s", err.Error())
		}

	} else {

		// Verify if the topic owner matches that of msg from
		if topic.Owner != msg.From {
			return nil, types.ErrSenderNotOwner.Wrapf(
				"from address %s doesn't match topic owner %s",
				msg.From, topic.Owner,
			)
		}

		// Verify order is the next expected order for the topic
		nextTopicOrder := topic.Order.Add(math.OneInt())
		if !nextTopicOrder.Equal(msg.Order) {
			return nil, types.ErrOrderNotMatching.Wrapf(
				"msg order %s doesn't match next topic order %s",
				msg.Order.String(), nextTopicOrder.String(),
			)
		}

		// Increment the topic order
		topic.Order = nextTopicOrder
	}

	// Set the topic
	k.SetTopic(ctx, topic)

	// Update the ethereum nonce
	nonce := k.bridgeKeeper.MustGetLastEthereumNonce(ctx).AddRaw(1)
	k.bridgeKeeper.SetLastEthereumNonce(ctx, nonce)

	// Addresses are lowercase for simpler parsing on Ethereum.
	return &types.MsgPostBlobResponse{
		Nonce: nonce,
		From:  strings.ToLower(msg.From),
		Topic: msg.Topic,
		Order: msg.Order,
		Data:  msg.Data,
	}, nil
}
