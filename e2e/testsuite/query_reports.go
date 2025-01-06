package testsuite

import (
	"context"

	reportstypes "github.com/fuel-infrastructure/fuel-sequencer/x/reports/types"
)

func (s *E2ETestSuite) QuerySlashReport(ctx context.Context, height uint64) *reportstypes.SlashReport {
	queryClient := s.getGRPCClients().ReportsQueryClient
	res, err := queryClient.SlashReport(ctx, &reportstypes.QueryGetSlashReportRequest{Height: height})
	s.Require().NoError(err)

	return &res.SlashReport
}
