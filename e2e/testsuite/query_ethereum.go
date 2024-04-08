package testsuite

import "context"

func (s *E2ETestSuite) GetEthereumHeight(ctx context.Context) (uint64, error) {
	return s.Chain.EthereumHeight(ctx)
}
