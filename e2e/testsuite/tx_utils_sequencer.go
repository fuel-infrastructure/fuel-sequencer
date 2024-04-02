package testsuite

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec/unknownproto"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
)

func decodeTx(txBytes []byte) (*sdktx.Tx, error) {
	var raw sdktx.TxRaw

	// reject all unknown proto fields in the root TxRaw
	err := unknownproto.RejectUnknownFieldsStrict(txBytes, &raw, encodingConfig.InterfaceRegistry)
	if err != nil {
		return nil, fmt.Errorf("failed to reject unknown fields: %w", err)
	}

	if err := cdc.Unmarshal(txBytes, &raw); err != nil {
		return nil, err
	}

	var body sdktx.TxBody
	if err := cdc.Unmarshal(raw.BodyBytes, &body); err != nil {
		return nil, fmt.Errorf("failed to decode tx: %w", err)
	}

	var authInfo sdktx.AuthInfo

	// reject all unknown proto fields in AuthInfo
	err = unknownproto.RejectUnknownFieldsStrict(raw.AuthInfoBytes, &authInfo, encodingConfig.InterfaceRegistry)
	if err != nil {
		return nil, fmt.Errorf("failed to reject unknown fields: %w", err)
	}

	if err := cdc.Unmarshal(raw.AuthInfoBytes, &authInfo); err != nil {
		return nil, fmt.Errorf("failed to decode auth info: %w", err)
	}

	return &sdktx.Tx{
		Body:       &body,
		AuthInfo:   &authInfo,
		Signatures: raw.Signatures,
	}, nil
}

// AssertValidTxResponse verifies that an sdk.TxResponse has non-empty values.
func (s *E2ETestSuite) AssertValidTxResponse(resp sdk.TxResponse) {
	errorMsg := fmt.Sprintf("%+v", resp)
	s.Require().NotEmpty(resp.TxHash, errorMsg)
	s.Require().Zero(resp.Code, errorMsg)
}

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
