package testsuite

import (
	"bytes"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec/unknownproto"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
)

func decodeTx(txBytes []byte) (*sdktx.Tx, error) {
	var raw sdktx.TxRaw

	// reject all unknown proto fields in the root TxRaw
	err := unknownproto.RejectUnknownFieldsStrict(txBytes, &raw, encodingConfig.InterfaceRegistry)
	if err != nil {
		return nil, fmt.Errorf("failed to reject unknown fields: %w", err)
	}

	if err := Cdc.Unmarshal(txBytes, &raw); err != nil {
		return nil, err
	}

	var body sdktx.TxBody
	if err := Cdc.Unmarshal(raw.BodyBytes, &body); err != nil {
		return nil, fmt.Errorf("failed to decode tx: %w", err)
	}

	var authInfo sdktx.AuthInfo

	// reject all unknown proto fields in AuthInfo
	err = unknownproto.RejectUnknownFieldsStrict(raw.AuthInfoBytes, &authInfo, encodingConfig.InterfaceRegistry)
	if err != nil {
		return nil, fmt.Errorf("failed to reject unknown fields: %w", err)
	}

	if err := Cdc.Unmarshal(raw.AuthInfoBytes, &authInfo); err != nil {
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
	return s.SubmitMsgsFrom(s.Chain.Validators[0], msgs...)
}

func (s *E2ETestSuite) SubmitMsgsWithGas(gas uint64, msgs ...sdk.Msg) (*sdk.TxResponse, error) {
	return s.SubmitMsgsWithGasFrom(s.Chain.Validators[0], gas, msgs...)
}

func (s *E2ETestSuite) SubmitMsgsFrom(val *validator, msgs ...sdk.Msg) (*sdk.TxResponse, error) {
	return s.SubmitMsgsWithGasFrom(val, defaultTxGas, msgs...)
}

func (s *E2ETestSuite) SubmitMsgsFromValidatorN(i int, msgs ...sdk.Msg) (*sdk.TxResponse, error) {
	return s.SubmitMsgsWithGasFrom(s.Chain.validators[i], defaultTxGas, msgs...)
}

func (s *E2ETestSuite) SubmitMsgsWithGasFrom(val *validator, gas uint64, msgs ...sdk.Msg) (*sdk.TxResponse, error) {

	kr, err := val.keyring()
	s.Require().NoError(err)
	addr := val.hostRPCPort

	outputBuffer := &bytes.Buffer{} // TODO: consider reusing this buffer with a reset in between each use
	clientCtx, err := s.Chain.clientContext(addr, &kr, validatorKeyName, val.Address(), outputBuffer)
	s.Require().NoError(err)

	respWithTxHash, err := s.Chain.sendMsgs(*clientCtx, outputBuffer, gas, msgs...)
	s.Require().NoError(err)

	var resp *sdk.TxResponse
	err = WaitForCondition(time.Second*30, time.Millisecond*500, func() (bool, error) {
		resp, err = authtx.QueryTx(*clientCtx, respWithTxHash.TxHash)
		if err != nil {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		// If we fail to query the tx, it means an error occurred with the original message broadcast.
		// We can return the original response instead.
		return respWithTxHash, nil
	}

	return resp, nil
}
