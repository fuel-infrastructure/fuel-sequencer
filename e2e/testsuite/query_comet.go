package testsuite

import (
	"context"
	"sort"

	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
)

func (s *E2ETestSuite) GetFuelSequencerHeight(ctx context.Context) (uint64, error) {
	return s.chain.FuelSequencerHeight(ctx)
}

// GetBlockByHeight fetches the block at a given height. Note: we are explicitly using the res.Block type which has been
// deprecated instead of res.SdkBlock to support backwards compatibility tests.
func (s *E2ETestSuite) GetBlockByHeight(ctx context.Context, height uint64) (*cmtservice.Block, error) {
	tmService := s.getGRPCClients().ConsensusServiceClient
	res, err := tmService.GetBlockByHeight(ctx, &cmtservice.GetBlockByHeightRequest{
		Height: int64(height),
	})
	if err != nil {
		return nil, err
	}

	return res.SdkBlock, nil
}

// GetValidatorSetByHeight returns the validators of the given chain at the specified height. The returned validators
// are sorted by address.
func (s *E2ETestSuite) GetValidatorSetByHeight(ctx context.Context, height uint64) ([]*cmtservice.Validator, error) {
	tmService := s.getGRPCClients().ConsensusServiceClient
	res, err := tmService.GetValidatorSetByHeight(ctx, &cmtservice.GetValidatorSetByHeightRequest{
		Height: int64(height),
	})
	if err != nil {
		return nil, err
	}

	sort.SliceStable(res.Validators, func(i, j int) bool {
		return res.Validators[i].Address < res.Validators[j].Address
	})

	return res.Validators, nil
}
