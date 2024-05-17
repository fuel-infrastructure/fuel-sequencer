package types_test

import (
	"testing"

	testtypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	"github.com/stretchr/testify/require"
)

func TestMsgDepositFromEthereum_ValidateBasic(t *testing.T) {
	tests := []struct {
		name      string
		msg       *types.MsgDepositFromEthereum
		expErrMsg string
	}{
		{
			name: "valid - recipient is a Hex address",
			msg:  testtypes.TestDepositEvent1.ToMsgDepositFromEthereum(testtypes.TestGovernanceAddress),
		},
		{
			name: "valid - recipient is a Sequencer address",
			msg:  testtypes.TestDepositEvent9.ToMsgDepositFromEthereum(testtypes.TestGovernanceAddress),
		},
		{
			name: "valid - recipient is the null address",
			msg:  testtypes.TestDepositEvent2.ToMsgDepositFromEthereum(testtypes.TestGovernanceAddress),
		},
		{
			name:      "nil receiver - error",
			msg:       nil,
			expErrMsg: "MsgDepositFromEthereum is nil",
		},
		{
			name: "invalid depositor - NO ERROR",
			msg: &types.MsgDepositFromEthereum{
				Depositor: "invalid-depositor",
				Recipient: testtypes.TestTo1,
				Amount:    testtypes.TestAmount1,
				Lockup:    testtypes.TestLockup1,
			},
		},
		{
			name: "invalid recipient - NO ERROR",
			msg: &types.MsgDepositFromEthereum{
				Depositor: testtypes.TestFrom1,
				Recipient: "invalid-recipient",
				Amount:    testtypes.TestAmount1,
				Lockup:    testtypes.TestLockup1,
			},
		},
		{
			name: "invalid lockup - error",
			msg: &types.MsgDepositFromEthereum{
				Depositor: testtypes.TestFrom1,
				Recipient: testtypes.TestTo1,
				Amount:    testtypes.TestAmount1,
				Lockup:    "0.23523",
			},
			expErrMsg: "could not convert lockup to a valid sdk.Int",
		},
		{
			name: "amount is zero - error",
			msg: &types.MsgDepositFromEthereum{
				Depositor: testtypes.TestFrom1,
				Recipient: testtypes.TestTo1,
				Amount:    "0",
				Lockup:    testtypes.TestLockup1,
			},
			expErrMsg: "amount must be bigger than zero",
		},
		{
			name: "amount is float - error",
			msg: &types.MsgDepositFromEthereum{
				Depositor: testtypes.TestFrom1,
				Recipient: testtypes.TestTo1,
				Amount:    "0.4356346",
				Lockup:    testtypes.TestLockup1,
			},
			expErrMsg: "could not convert amount to a valid sdk.Int",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.expErrMsg != "" {
				require.ErrorContains(t, err, tt.expErrMsg)
				return
			}
			require.NoError(t, err)
		})
	}
}
