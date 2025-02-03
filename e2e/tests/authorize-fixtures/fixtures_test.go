package authorize_fixtures_test

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"

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
	authzMsgTypeUrl := "/fuelsequencer.bridge.v1.MsgWithdrawToEthereum"
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

	var logsString string

	// Deposit
	data := testsuite.PackDeposit(amount)
	tx, err := s.SendEthTransactionToSequencerInterfaceContract(data)
	s.Require().NoError(err)
	logs, err := json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logsString += fmt.Sprintf("DepositLogs = `%s`\n", logs)

	// DepositFor
	data = testsuite.PackDepositFor(amount, receiverAddressEth)
	tx, err = s.SendEthTransactionToSequencerInterfaceContract(data)
	s.Require().NoError(err)
	logs, err = json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logsString += fmt.Sprintf("DepositForLogs = `%s`\n", logs)

	// Deposit with lockup
	vestingDuration := testsuite.VestingDuration2Years
	amountScaledDown := new(big.Int).Quo(amount, testsuite.MigrationRatio)
	tx = s.DepositTokenToSequencerFromMigrationNoDelegation(amountScaledDown, vestingDuration)
	s.Require().NoError(err)
	logs, err = json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logsString += fmt.Sprintf("DepositWithLockupLogs = `%s`\n", logs)

	// Delegate
	data = testsuite.PackDelegate(amount, validator1AddressEth)
	tx, err = s.SendEthTransactionToSequencerInterfaceContract(data)
	s.Require().NoError(err)
	logs, err = json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logsString += fmt.Sprintf("DelegateLogs = `%s`\n", logs)

	// Redelegate
	data = testsuite.PackRedelegate(amount, validator1AddressEth, validator2AddressEth)
	tx, err = s.SendEthTransactionToSequencerInterfaceContract(data)
	s.Require().NoError(err)
	logs, err = json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logsString += fmt.Sprintf("RedelegateLogs = `%s`\n", logs)

	// ClaimRewards
	data = testsuite.PackClaimRewards(validator1AddressEth)
	tx, err = s.SendEthTransactionToSequencerInterfaceContract(data)
	s.Require().NoError(err)
	logs, err = json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logsString += fmt.Sprintf("ClaimRewardsLogs = `%s`\n", logs)

	// Unbond
	data = testsuite.PackUnbond(amount, validator1AddressEth)
	tx, err = s.SendEthTransactionToSequencerInterfaceContract(data)
	s.Require().NoError(err)
	logs, err = json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logsString += fmt.Sprintf("UnbondLogs = `%s`\n", logs)

	// Withdraw
	data = testsuite.PackWithdraw(amount)
	tx, err = s.SendEthTransactionToSequencerInterfaceContract(data)
	s.Require().NoError(err)
	logs, err = json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logsString += fmt.Sprintf("WithdrawLogs = `%s`\n", logs)

	// WithdrawTo
	data = testsuite.PackWithdrawTo(amount, receiverAddressEth)
	tx, err = s.SendEthTransactionToSequencerInterfaceContract(data)
	s.Require().NoError(err)
	logs, err = json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logsString += fmt.Sprintf("WithdrawToLogs = `%s`\n", logs)

	// Transfer
	data = testsuite.PackTransfer(receiverAddressEth, amount)
	tx, err = s.SendEthTransactionToSequencerInterfaceContract(data)
	s.Require().NoError(err)
	logs, err = json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logsString += fmt.Sprintf("TransferLogs = `%s`\n", logs)

	// Vote
	voteProposalId := uint64(1)
	voteOption := uint32(govtypesv1.VoteOption_VOTE_OPTION_YES)
	voteMetadata := "metadata"
	data = testsuite.PackVote(voteProposalId, voteOption, voteMetadata)
	tx, err = s.SendEthTransactionToSequencerInterfaceContract(data)
	s.Require().NoError(err)
	logs, err = json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logsString += fmt.Sprintf("VoteLogs = `%s`\n", logs)

	// SetRewardRecipient
	data = testsuite.PackSetRewardRecipient(receiverAddressEth)
	tx, err = s.SendEthTransactionToSequencerInterfaceContract(data)
	s.Require().NoError(err)
	logs, err = json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logsString += fmt.Sprintf("SetRewardRecipientLogs = `%s`\n", logs)

	// Grant with no expiration
	data = testsuite.PackGrant(receiverAddressEth, authzMsgTypeUrl, 0)
	tx, err = s.SendEthTransactionToSequencerInterfaceContract(data)
	s.Require().NoError(err)
	logs, err = json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logsString += fmt.Sprintf("GrantNoExpirationLogs = `%s`\n", logs)

	// Grant with expiration
	data = testsuite.PackGrant(receiverAddressEth, authzMsgTypeUrl, authzExpiration)
	tx, err = s.SendEthTransactionToSequencerInterfaceContract(data)
	s.Require().NoError(err)
	logs, err = json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logsString += fmt.Sprintf("GrantWithExpirationLogs = `%s`\n", logs)

	// Revoke
	data = testsuite.PackRevoke(receiverAddressEth, authzMsgTypeUrl)
	tx, err = s.SendEthTransactionToSequencerInterfaceContract(data)
	s.Require().NoError(err)
	logs, err = json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logsString += fmt.Sprintf("RevokeLogs = `%s`\n", logs)

	// Print the fixtures
	fmt.Printf("Amount = %d", amount.Uint64())
	fmt.Printf("\nBridgeDenom = \"%s\"", testsuite.BridgeDenom)
	fmt.Printf("\nVestingDurationSeconds = \"%d\"", int64(vestingDuration.Seconds()))
	fmt.Printf("\nVoteProposalId = uint64(%d)", voteProposalId)
	fmt.Printf("\nVoteOption = int32(%d)", voteOption)
	fmt.Printf("\nVoteMetadata = \"%s\"", voteMetadata)
	fmt.Printf("\nSenderAddress = \"%s\"", senderAddress)
	fmt.Printf("\nReceiverAddress = \"%s\"", receiverAddress)
	fmt.Printf("\nValidator1Address = \"%s\"", validator1AddressEth)
	fmt.Printf("\nValidator2Address = \"%s\"", validator2AddressEth)
	fmt.Printf("\nSequencerProxyContractAddress = \"%s\"", testsuite.SequencerProxyContractAddressStr)
	fmt.Printf("\nAuthzMsgTypeUrl = \"%s\"", authzMsgTypeUrl)
	fmt.Printf("\nAuthzExpiration = uint32(%d)", authzExpiration)
	fmt.Print("\n\n")
	fmt.Println(logsString)
}
