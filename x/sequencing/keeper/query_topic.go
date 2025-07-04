package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/types"
)

func (k Keeper) TopicAll(
	ctx context.Context,
	req *types.QueryAllTopicRequest,
) (*types.QueryAllTopicResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	var topics []types.Topic

	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.TopicKey))

	pageRes, err := query.Paginate(store, req.Pagination, func(key []byte, value []byte) error {
		var topic types.Topic
		if err := k.cdc.Unmarshal(value, &topic); err != nil {
			return err
		}

		topics = append(topics, topic)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllTopicResponse{Topic: topics, Pagination: pageRes}, nil
}

func (k Keeper) Topic(ctx context.Context, req *types.QueryGetTopicRequest) (*types.QueryGetTopicResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	val, found := k.GetTopic(
		ctx,
		req.Id,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetTopicResponse{Topic: val}, nil
}
