package utils

import (
	"encoding/json"
	"strconv"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ethereumtypes "github.com/ethereum/go-ethereum/core/types"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	"github.com/fuel-infrastructure/fuel-sequencer/sidecar/testutil"
	"github.com/fuel-infrastructure/fuel-sequencer/sidecar/testutil/fixtures"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractLogDataToEvent_AuthorizeTxFromEvent(t *testing.T) {
	sequencerProxyABI := testutil.SequencerProxyContractABI(t)
	amountParsed := sdkmath.NewInt(fixtures.Amount)

	testCases := []struct {
		name        string
		logs        string
		getExpEvent func() *sidecartypes.Event
		expErrMsg   string
	}{
		{
			name: "Deposit event via Deposit",
			logs: fixtures.DepositLogs,
			getExpEvent: func() *sidecartypes.Event {
				return testutil.EventFromDepositEvent(t, sidecartypes.DepositEvent{
					Depositor: fixtures.SenderAddress,
					Recipient: fixtures.SenderAddress, // sender deposits to themselves
					Amount:    strconv.FormatInt(fixtures.Amount, 10),
					Lockup:    "0",
				})
			},
		},
		{
			name: "Deposit event via DepositFor",
			logs: fixtures.DepositForLogs,
			getExpEvent: func() *sidecartypes.Event {
				return testutil.EventFromDepositEvent(t, sidecartypes.DepositEvent{
					Depositor: fixtures.SenderAddress,
					Recipient: fixtures.ReceiverAddress,
					Amount:    strconv.FormatInt(fixtures.Amount, 10),
					Lockup:    "0",
				})
			},
		},
		{
			name: "Deposit event with lockup",
			logs: fixtures.DepositWithLockupLogs,
			getExpEvent: func() *sidecartypes.Event {
				return testutil.EventFromDepositEvent(t, sidecartypes.DepositEvent{
					Depositor: fixtures.SenderAddress,
					Recipient: fixtures.SenderAddress, // sender deposits to themselves
					Amount:    strconv.FormatInt(fixtures.Amount, 10),
					Lockup:    fixtures.VestingDurationSeconds,
				})
			},
		},
		{
			name: "Delegate event",
			logs: fixtures.DelegateLogs,
			getExpEvent: func() *sidecartypes.Event {
				return testutil.EventFromMsg(t, &stakingtypes.MsgDelegate{
					DelegatorAddress: fixtures.SenderAddress,
					ValidatorAddress: fixtures.Validator1Address,
					Amount:           sdk.NewCoin(fixtures.BridgeDenom, amountParsed),
				})
			},
		},
		{
			name: "Redelegate event",
			logs: fixtures.RedelegateLogs,
			getExpEvent: func() *sidecartypes.Event {
				return testutil.EventFromMsg(t, &stakingtypes.MsgBeginRedelegate{
					DelegatorAddress:    fixtures.SenderAddress,
					ValidatorSrcAddress: fixtures.Validator1Address,
					ValidatorDstAddress: fixtures.Validator2Address,
					Amount:              sdk.NewCoin(fixtures.BridgeDenom, amountParsed),
				})
			},
		},
		{
			name: "ClaimRewards event",
			logs: fixtures.ClaimRewardsLogs,
			getExpEvent: func() *sidecartypes.Event {
				return testutil.EventFromMsg(t, &distributiontypes.MsgWithdrawDelegatorReward{
					DelegatorAddress: fixtures.SenderAddress,
					ValidatorAddress: fixtures.Validator1Address,
				})
			},
		},
		{
			name: "Unbond event",
			logs: fixtures.UnbondLogs,
			getExpEvent: func() *sidecartypes.Event {
				return testutil.EventFromMsg(t, &stakingtypes.MsgUndelegate{
					DelegatorAddress: fixtures.SenderAddress,
					ValidatorAddress: fixtures.Validator1Address,
					Amount:           sdk.NewCoin(fixtures.BridgeDenom, amountParsed),
				})
			},
		},
		{
			name: "Withdraw event via Withdraw",
			logs: fixtures.WithdrawLogs,
			getExpEvent: func() *sidecartypes.Event {
				return testutil.EventFromMsg(t, &bridgetypes.MsgWithdrawToEthereum{
					From:   fixtures.SenderAddress,
					To:     fixtures.SenderAddress, // sender withdraws to themselves
					Amount: sdk.NewCoin(fixtures.BridgeDenom, amountParsed),
				})
			},
		},
		{
			name: "Withdraw event via WithdrawTo",
			logs: fixtures.WithdrawToLogs,
			getExpEvent: func() *sidecartypes.Event {
				return testutil.EventFromMsg(t, &bridgetypes.MsgWithdrawToEthereum{
					From:   fixtures.SenderAddress,
					To:     fixtures.ReceiverAddress,
					Amount: sdk.NewCoin(fixtures.BridgeDenom, amountParsed),
				})
			},
		},
		{
			name: "Transfer event",
			logs: fixtures.TransferLogs,
			getExpEvent: func() *sidecartypes.Event {
				return testutil.EventFromMsg(t, &banktypes.MsgSend{
					FromAddress: fixtures.SenderAddress,
					ToAddress:   fixtures.ReceiverAddress,
					Amount:      sdk.NewCoins(sdk.NewCoin(fixtures.BridgeDenom, amountParsed)),
				})
			},
		},
		{
			name: "Vote event",
			logs: fixtures.VoteLogs,
			getExpEvent: func() *sidecartypes.Event {
				return testutil.EventFromMsg(t, &govtypesv1.MsgVote{
					ProposalId: fixtures.VoteProposalId,
					Voter:      fixtures.SenderAddress,
					Option:     govtypesv1.VoteOption(fixtures.VoteOption),
					Metadata:   fixtures.VoteMetadata,
				})
			},
		},
		{
			name: "SetRewardRecipient event",
			logs: fixtures.SetRewardRecipientLogs,
			getExpEvent: func() *sidecartypes.Event {
				return testutil.EventFromMsg(t, &distributiontypes.MsgSetWithdrawAddress{
					DelegatorAddress: fixtures.SenderAddress,
					WithdrawAddress:  fixtures.ReceiverAddress,
				})
			},
		},
		{
			name: "Grant Claim Rewards event with no expiration",
			logs: fixtures.GrantClaimRewardsNoExpirationLogs,
			getExpEvent: func() *sidecartypes.Event {
				msgGrant := authz.MsgGrant{
					Granter: fixtures.SenderAddress,
					Grantee: fixtures.ReceiverAddress,
					Grant: authz.Grant{
						Expiration: nil,
					},
				}
				err := msgGrant.SetAuthorization(authz.NewGenericAuthorization("/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward"))
				if err != nil {
					t.Fatalf("error when setting authorization: %x", err)
				}

				return testutil.EventFromMsg(t, &msgGrant)
			},
		},
		{
			name: "Grant Claim Rewards event with expiration",
			logs: fixtures.GrantClaimRewardsWithExpirationLogs,
			getExpEvent: func() *sidecartypes.Event {
				if fixtures.AuthzExpiration == 0 {
					t.Fatalf("AuthzExpiration is 0 - use a non-zero value to confirm setting the expiration works...")
				}
				expiration := time.Unix(int64(fixtures.AuthzExpiration), 0)

				msgGrant := authz.MsgGrant{
					Granter: fixtures.SenderAddress,
					Grantee: fixtures.ReceiverAddress,
					Grant: authz.Grant{
						Expiration: &expiration,
					},
				}
				err := msgGrant.SetAuthorization(authz.NewGenericAuthorization("/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward"))
				if err != nil {
					t.Fatalf("error when setting authorization: %x", err)
				}

				return testutil.EventFromMsg(t, &msgGrant)
			},
		},
		{
			name: "Revoke Claim Rewards event",
			logs: fixtures.RevokeClaimRewardsLogs,
			getExpEvent: func() *sidecartypes.Event {
				return testutil.EventFromMsg(t, &authz.MsgRevoke{
					Granter:    fixtures.SenderAddress,
					Grantee:    fixtures.ReceiverAddress,
					MsgTypeUrl: "/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward",
				})
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expectedEvent := tc.getExpEvent()

			// Parse log into Log struct
			var parsedLog []ethereumtypes.Log
			err := json.Unmarshal([]byte(tc.logs), &parsedLog)
			require.NoError(t, err)

			// There might be multiple logs. Check that there is exactly one log that matches the expected log.
			matches := 0
			for _, log := range parsedLog {
				event, err := ExtractLogDataToEvent(log, sequencerProxyABI, fixtures.BridgeDenom)
				if tc.expErrMsg != "" {
					require.EqualError(t, err, tc.expErrMsg)
					return
				}
				require.NoError(t, err)

				if assert.ObjectsAreEqual(expectedEvent, event) {
					matches += 1
				}
			}

			require.Equal(t, 1, matches)
		})
	}
}
