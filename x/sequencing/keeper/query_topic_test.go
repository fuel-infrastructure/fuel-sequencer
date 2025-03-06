package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/fuel-infrastructure/fuel-sequencer/testutil"

	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/testutil/nullify"
	"github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/types"
)

func TestQueryTopicAll(t *testing.T) {
	keeper, ctx := keepertest.SequencingKeeper(t)

	testTopicOne := types.Topic{
		Id:    testutil.MockTopicIDHex(1),
		Owner: "ownerAddress",
		Order: math.ZeroInt(),
	}
	testTopicTwo := types.Topic{
		Id:    testutil.MockTopicIDHex(2),
		Owner: "ownerAddress",
		Order: math.ZeroInt(),
	}
	keeper.SetTopic(ctx, testTopicOne)
	keeper.SetTopic(ctx, testTopicTwo)

	tests := []struct {
		desc     string
		request  *types.QueryAllTopicRequest
		response *types.QueryAllTopicResponse
		err      error
	}{
		{
			desc:     "ValidRequest",
			request:  &types.QueryAllTopicRequest{},
			response: &types.QueryAllTopicResponse{Topic: []types.Topic{testTopicOne, testTopicTwo}, Pagination: nil},
		},
		{
			desc:    "InvalidRequest",
			request: nil,
			err:     status.Error(codes.InvalidArgument, "request cannot be nil"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.TopicAll(ctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.Equal(t,
					nullify.Fill(tc.response.Topic),
					nullify.Fill(response.Topic),
				)
			}
		})
	}
}

func TestQueryTopic(t *testing.T) {
	keeper, ctx := keepertest.SequencingKeeper(t)
	testTopic := types.Topic{
		Id:    testutil.MockTopicIDHex(1),
		Owner: "ownerAddress",
		Order: math.ZeroInt(),
	}
	keeper.SetTopic(ctx, testTopic)

	tests := []struct {
		desc     string
		request  *types.QueryGetTopicRequest
		response *types.QueryGetTopicResponse
		err      error
	}{
		{
			desc:     "ValidRequest",
			request:  &types.QueryGetTopicRequest{Id: testutil.MockTopicIDHex(1)},
			response: &types.QueryGetTopicResponse{Topic: testTopic},
		},
		{
			desc:    "InvalidRequest",
			request: nil,
			err:     status.Error(codes.InvalidArgument, "request cannot be nil"),
		},
		{
			desc:    "NotFound",
			request: &types.QueryGetTopicRequest{Id: testutil.MockTopicIDHex(9999)},
			err:     status.Error(codes.NotFound, "not found"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.Topic(ctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.Equal(t,
					nullify.Fill(tc.response),
					nullify.Fill(response),
				)
			}
		})
	}
}
