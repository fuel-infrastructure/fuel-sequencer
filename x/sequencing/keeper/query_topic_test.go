package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/testutil/nullify"
	"github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/types"
)

func TestQueryTopicAll(t *testing.T) {
	keeper, ctx := keepertest.SequencingKeeper(t)

	testTopicOne := types.Topic{
		Id:    math.NewInt(1),
		Owner: "ownerAddress",
		Order: math.ZeroInt(),
	}
	testTopicTwo := types.Topic{
		Id:    math.NewInt(2),
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
		Id:    math.NewInt(1),
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
			request:  &types.QueryGetTopicRequest{Id: "1"},
			response: &types.QueryGetTopicResponse{Topic: testTopic},
		},
		{
			desc:    "InvalidRequest",
			request: nil,
			err:     status.Error(codes.InvalidArgument, "request cannot be nil"),
		},
		{
			desc:    "InvalidRequest_NoId",
			request: &types.QueryGetTopicRequest{Id: ""},
			err:     status.Error(codes.InvalidArgument, "invalid request"),
		},
		{
			desc:    "NotFound",
			request: &types.QueryGetTopicRequest{Id: "9999"},
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

func TestQueryNextTopicId(t *testing.T) {
	keeper, ctx := keepertest.SequencingKeeper(t)

	nextTopicId := math.NewInt(3)
	keeper.SetNextTopicId(ctx, nextTopicId)

	tests := []struct {
		desc     string
		request  *types.QueryGetNextTopicIdRequest
		response *types.QueryGetNextTopicIdResponse
		err      error
	}{
		{
			desc:     "ValidRequest",
			request:  &types.QueryGetNextTopicIdRequest{},
			response: &types.QueryGetNextTopicIdResponse{NextTopicId: nextTopicId.String()},
		},
		{
			desc:    "InvalidRequest",
			request: nil,
			err:     status.Error(codes.InvalidArgument, "request cannot be nil"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.NextTopicId(ctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.response, response)
			}
		})
	}
}
