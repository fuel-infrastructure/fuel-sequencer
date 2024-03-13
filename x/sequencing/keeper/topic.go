package keeper

import (
	"context"
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/types"
)

// SetTopic set a specific topic in the store from its index.
func (k Keeper) SetTopic(ctx context.Context, topic types.Topic) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.TopicKeyPrefix))
	b := k.cdc.MustMarshal(&topic)
	store.Set(types.TopicKey(
		topic.Id.String(),
	), b)
}

// GetTopic returns a topic from its index.
func (k Keeper) GetTopic(
	ctx context.Context,
	index math.Int,
) (val types.Topic, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.TopicKeyPrefix))

	b := store.Get(types.TopicKey(
		index.String(),
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveTopic removes a topic from the store
func (k Keeper) RemoveTopic(
	ctx context.Context,
	index math.Int,
) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.TopicKeyPrefix))
	store.Delete(types.TopicKey(
		index.String(),
	))
}

// GetAllTopic returns all topics.
func (k Keeper) GetAllTopic(ctx context.Context) (list []types.Topic) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.TopicKeyPrefix))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Topic
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

// HasTopic checks if the topic exists in the store.
func (k Keeper) HasTopic(ctx context.Context, topicId string) bool {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.TopicKeyPrefix))
	topicKey := types.TopicKey(topicId)
	return store.Has(topicKey)
}

// SetNextTopicId sets next topic Id.
func (k Keeper) SetNextTopicId(ctx context.Context, topicId math.Int) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.NextGlobalTopicIdKey)
	b := topicId.BigInt().Bytes()
	store.Set([]byte{0}, b)
}

// MustGetNextTopicId returns the next topic id, panics if the global topic id has not been initialized.
func (k Keeper) MustGetNextTopicId(ctx context.Context) (val math.Int) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.NextGlobalTopicIdKey)

	b := store.Get([]byte{0})
	if b == nil {
		panic(fmt.Errorf("next topic id not found"))
	}

	val = math.NewIntFromBigInt(new(big.Int).SetBytes(b))

	return val
}
