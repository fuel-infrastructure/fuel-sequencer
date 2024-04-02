package testsuite

import sdk "github.com/cosmos/cosmos-sdk/types"

func (s *E2ETestSuite) SubmitMsgs(msgs ...sdk.Msg) (*sdk.TxResponse, error) {
	val := s.Chain.validators[0]

	kr, err := val.keyring()
	s.Require().NoError(err)
	addr := val.hostRPCPort

	clientCtx, err := s.Chain.clientContext(addr, &kr, validatorKeyName, val.address())
	s.Require().NoError(err)

	return s.Chain.sendMsgs(*clientCtx, msgs...)
}
