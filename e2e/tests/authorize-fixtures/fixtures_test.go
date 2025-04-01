package authorize_fixtures_test

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strings"

	sdkmath "cosmossdk.io/math"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
)

// TestGenerateEventLogFixtures generates fixtures used in unit testing.
// The generated fixtures should be copied to sidecar/testutil/fixtures/fixtures.go
// They are currently only used by sidecar/utils/utils.go.
func (s *AuthorizeFixturesTestSuite) TestGenerateEventLogFixtures() {
	senderAddress := s.EthKeys[0].AddressHex
	receiverAddress := s.EthKeys[1].AddressHex
	receiverAddressEth := s.EthKeys[1].Address
	validator1AddressEth := s.SeqKeys[0].ValAddressEth
	validator2AddressEth := s.SeqKeys[1].ValAddressEth
	authzExpiration := uint32(math.MaxUint32) // Unix Timestamp about year 2106
	// Approve V2 tokens for use by SequencerInterfaceContract.
	approveAmount := new(big.Int).SetInt64(1000000000000)
	approveData := testsuite.PackApproveToken(testsuite.SequencerInterfaceContractAddress, approveAmount)
	_, err := s.SendEthTransactionToTokenContract(approveData)
	s.Require().NoError(err)

	// Amount to be used whenever an amount is needed. Since we do a deposit via migrate, this value should be greater
	// or equal to the migration ratio so that we do not end up with decimal amounts, which cause the test to fail.
	amountSDK := sdkmath.NewInt(100)
	amount := amountSDK.BigInt()

	var constLines, logLines []string

	// Deposit
	data := testsuite.PackDeposit(amount)
	tx, err := s.SendEthTransactionToSequencerInterfaceContract(data)
	s.Require().NoError(err)
	logs, err := json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logLines = append(logLines, fmt.Sprintf("DepositLogs = `%s`", logs))

	// DepositFor
	data = testsuite.PackDepositFor(amount, receiverAddressEth)
	tx, err = s.SendEthTransactionToSequencerInterfaceContract(data)
	s.Require().NoError(err)
	logs, err = json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logLines = append(logLines, fmt.Sprintf("DepositForLogs = `%s`", logs))

	// Deposit with lockup
	vestingDuration := testsuite.VestingDuration2Years
	amountScaledDown := new(big.Int).Quo(amount, testsuite.MigrationRatio)
	tx = s.DepositTokenToSequencerFromMigrationNoDelegation(amountScaledDown, vestingDuration)
	s.Require().NoError(err)
	logs, err = json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logLines = append(logLines, fmt.Sprintf("DepositWithLockupLogs = `%s`", logs))

	// Delegate
	data = testsuite.PackDelegate(amount, validator1AddressEth)
	tx, err = s.SendEthTransactionToSequencerInterfaceContract(data)
	s.Require().NoError(err)
	logs, err = json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logLines = append(logLines, fmt.Sprintf("DelegateLogs = `%s`", logs))

	// Redelegate
	data = testsuite.PackRedelegate(amount, validator1AddressEth, validator2AddressEth)
	tx, err = s.SendEthTransactionToSequencerInterfaceContract(data)
	s.Require().NoError(err)
	logs, err = json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logLines = append(logLines, fmt.Sprintf("RedelegateLogs = `%s`", logs))

	// ClaimRewards
	data = testsuite.PackClaimRewards(validator1AddressEth)
	tx, err = s.SendEthTransactionToSequencerInterfaceContract(data)
	s.Require().NoError(err)
	logs, err = json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logLines = append(logLines, fmt.Sprintf("ClaimRewardsLogs = `%s`", logs))

	// Unbond
	data = testsuite.PackUnbond(amount, validator1AddressEth)
	tx, err = s.SendEthTransactionToSequencerInterfaceContract(data)
	s.Require().NoError(err)
	logs, err = json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logLines = append(logLines, fmt.Sprintf("UnbondLogs = `%s`", logs))

	// Withdraw
	data = testsuite.PackWithdraw(amount)
	tx, err = s.SendEthTransactionToSequencerInterfaceContract(data)
	s.Require().NoError(err)
	logs, err = json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logLines = append(logLines, fmt.Sprintf("WithdrawLogs = `%s`", logs))

	// WithdrawTo
	data = testsuite.PackWithdrawTo(amount, receiverAddressEth)
	tx, err = s.SendEthTransactionToSequencerInterfaceContract(data)
	s.Require().NoError(err)
	logs, err = json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logLines = append(logLines, fmt.Sprintf("WithdrawToLogs = `%s`", logs))

	// Transfer
	data = testsuite.PackTransfer(receiverAddressEth, amount)
	tx, err = s.SendEthTransactionToSequencerInterfaceContract(data)
	s.Require().NoError(err)
	logs, err = json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logLines = append(logLines, fmt.Sprintf("TransferLogs = `%s`", logs))

	// Vote
	voteProposalId := uint64(1)
	voteOption := uint32(govtypesv1.VoteOption_VOTE_OPTION_YES)
	voteMetadata := "metadata"
	data = testsuite.PackVote(voteProposalId, voteOption, voteMetadata)
	tx, err = s.SendEthTransactionToSequencerInterfaceContract(data)
	s.Require().NoError(err)
	logs, err = json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logLines = append(logLines, fmt.Sprintf("VoteLogs = `%s`", logs))

	// SetRewardRecipient
	data = testsuite.PackSetRewardRecipient(receiverAddressEth)
	tx, err = s.SendEthTransactionToSequencerInterfaceContract(data)
	s.Require().NoError(err)
	logs, err = json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logLines = append(logLines, fmt.Sprintf("SetRewardRecipientLogs = `%s`", logs))

	// Grant Claim Rewards with no expiration
	data = testsuite.PackGrantClaimRewards(receiverAddressEth, 0)
	tx, err = s.SendEthTransactionToSequencerInterfaceContract(data)
	s.Require().NoError(err)
	logs, err = json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logLines = append(logLines, fmt.Sprintf("GrantClaimRewardsNoExpirationLogs = `%s`", logs))

	// Grant Claim Rewards with expiration
	data = testsuite.PackGrantClaimRewards(receiverAddressEth, authzExpiration)
	tx, err = s.SendEthTransactionToSequencerInterfaceContract(data)
	s.Require().NoError(err)
	logs, err = json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logLines = append(logLines, fmt.Sprintf("GrantClaimRewardsWithExpirationLogs = `%s`", logs))

	// Revoke Claim Rewards
	data = testsuite.PackRevokeClaimRewards(receiverAddressEth)
	tx, err = s.SendEthTransactionToSequencerInterfaceContract(data)
	s.Require().NoError(err)
	logs, err = json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logLines = append(logLines, fmt.Sprintf("RevokeClaimRewardsLogs = `%s`", logs))

	// Print the fixtures
	constLines = append(constLines,
		fmt.Sprintf("Amount = %d", amount.Uint64()),
		fmt.Sprintf("BridgeDenom = \"%s\"", testsuite.BridgeDenom),
		fmt.Sprintf("VestingDurationSeconds = \"%d\"", int64(vestingDuration.Seconds())),
		fmt.Sprintf("VoteProposalId = uint64(%d)", voteProposalId),
		fmt.Sprintf("VoteOption = int32(%d)", voteOption),
		fmt.Sprintf("VoteMetadata = \"%s\"", voteMetadata),
		fmt.Sprintf("SenderAddress = \"%s\"", senderAddress),
		fmt.Sprintf("ReceiverAddress = \"%s\"", receiverAddress),
		fmt.Sprintf("Validator1Address = \"%s\"", validator1AddressEth),
		fmt.Sprintf("Validator2Address = \"%s\"", validator2AddressEth),
		fmt.Sprintf("SequencerProxyContractAddress = \"%s\"", testsuite.SequencerProxyContractAddressStr),
		fmt.Sprintf("AuthzExpiration = uint32(%d)", authzExpiration),
	)

	constString := strings.Join(constLines, "\n\t")
	logString := strings.Join(logLines, "\n\t")

	template := `package fixtures

const (
	// Fixtures generated by running the test in e2e/tests/authorize-fixtures
	%s

	%s
)
	
`
	fmt.Printf(template, constString, logString)
}
