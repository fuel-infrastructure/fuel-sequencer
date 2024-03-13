package keeper

import (
	"context"
	"math/big"

	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.TopicKeyPrefix))

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

	Id, ok := math.NewIntFromString(req.Id)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	val, found := k.GetTopic(
		ctx,
		Id,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetTopicResponse{Topic: val}, nil
}

func (k Keeper) NextTopicId(
	ctx context.Context,
	req *types.QueryGetNextTopicIdRequest,
) (*types.QueryGetNextTopicIdResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.NextGlobalTopicIdKey)

	b := store.Get([]byte{0})
	if b == nil {
		return nil, status.Error(codes.NotFound, "next topic id not found")
	}

	nextTopicId := math.NewIntFromBigInt(new(big.Int).SetBytes(b))
	return &types.QueryGetNextTopicIdResponse{
		NextTopicId: nextTopicId.String(),
	}, nil
}
