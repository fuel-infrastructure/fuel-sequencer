package testsuite

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *E2ETestSuite) SubmitMsgs(msgs ...sdk.Msg) (*sdk.TxResponse, error) {
	return s.SubmitMsgsFrom(s.chain.validators[0], msgs...)
}

func (s *E2ETestSuite) SubmitMsgsFrom(val *validator, msgs ...sdk.Msg) (*sdk.TxResponse, error) {

	kr, err := val.keyring()
	s.Require().NoError(err)
	addr := val.hostRPCPort

	outputBuffer := &bytes.Buffer{} // TODO: consider reusing this buffer with a reset in between each use
	clientCtx, err := s.chain.clientContext(addr, &kr, validatorKeyName, val.address(), outputBuffer)
	s.Require().NoError(err)

	return s.chain.sendMsgs(*clientCtx, outputBuffer, msgs...)
}
