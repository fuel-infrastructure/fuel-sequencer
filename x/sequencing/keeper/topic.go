package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/types"
)

// SetTopic set a specific topic in the store by its Id.
func (k Keeper) SetTopic(ctx context.Context, topic types.Topic) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.TopicKeyPrefix))
	b := k.cdc.MustMarshal(&topic)
	store.Set(types.TopicKey(topic.Id), b)
}

// GetTopic returns a topic by its Id.
func (k Keeper) GetTopic(
	ctx context.Context,
	topicId []byte,
) (val types.Topic, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.TopicKeyPrefix))

	b := store.Get(types.TopicKey(topicId))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveTopic removes a topic from the store.
func (k Keeper) RemoveTopic(
	ctx context.Context,
	topicId []byte,
) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.TopicKeyPrefix))
	store.Delete(types.TopicKey(topicId))
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
func (k Keeper) HasTopic(ctx context.Context, topicId []byte) bool {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.TopicKeyPrefix))
	topicKey := types.TopicKey(topicId)
	return store.Has(topicKey)
}
