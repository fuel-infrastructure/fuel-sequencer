package authorize_fixtures_test

import (
	"encoding/json"
	"fmt"
	"math/big"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
)

// TestGenerateEventLogFixtures generates fixtures used in unit testing.
// The generated fixtures should be copied to testutil/fixtures/fixtures.go
// They are currently only used by sidecar/utils/utils.go.
func (s *AuthorizeFixturesTestSuite) TestGenerateEventLogFixtures() {
	senderAddress := s.EthKeys[0].AddressHex
	receiverAddress := s.EthKeys[1].AddressHex
	receiverAddressEth := s.EthKeys[1].Address
	validator1AddressEth := s.SeqKeys[0].ValAddressEth
	validator2AddressEth := s.SeqKeys[1].ValAddressEth

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
	logsString += fmt.Sprintf("depositLogs = `%s`\n", logs)

	// DepositFor
	data = testsuite.PackDepositFor(amount, receiverAddressEth)
	tx, err = s.SendEthTransactionToSequencerInterfaceContract(data)
	s.Require().NoError(err)
	logs, err = json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logsString += fmt.Sprintf("depositForLogs = `%s`\n", logs)

	// Deposit with lockup
	vestingDuration := testsuite.VestingDuration2Years
	amountScaledDown := new(big.Int).Quo(amount, testsuite.MigrationRatio)
	tx = s.DepositTokenToSequencerFromMigrationNoDelegation(amountScaledDown, vestingDuration)
	s.Require().NoError(err)
	logs, err = json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logsString += fmt.Sprintf("depositWithLockupLogs = `%s`\n", logs)

	// Delegate
	data = testsuite.PackDelegate(amount, validator1AddressEth)
	tx, err = s.SendEthTransactionToSequencerInterfaceContract(data)
	s.Require().NoError(err)
	logs, err = json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logsString += fmt.Sprintf("delegateLogs = `%s`\n", logs)

	// Redelegate
	data = testsuite.PackRedelegate(amount, validator1AddressEth, validator2AddressEth)
	tx, err = s.SendEthTransactionToSequencerInterfaceContract(data)
	s.Require().NoError(err)
	logs, err = json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logsString += fmt.Sprintf("redelegateLogs = `%s`\n", logs)

	// ClaimRewards
	data = testsuite.PackClaimRewards(validator1AddressEth)
	tx, err = s.SendEthTransactionToSequencerInterfaceContract(data)
	s.Require().NoError(err)
	logs, err = json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logsString += fmt.Sprintf("claimRewardsLogs = `%s`\n", logs)

	// Unbond
	data = testsuite.PackUnbond(amount, validator1AddressEth)
	tx, err = s.SendEthTransactionToSequencerInterfaceContract(data)
	s.Require().NoError(err)
	logs, err = json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logsString += fmt.Sprintf("unbondLogs = `%s`\n", logs)

	// Withdraw
	data = testsuite.PackWithdraw(amount)
	tx, err = s.SendEthTransactionToSequencerInterfaceContract(data)
	s.Require().NoError(err)
	logs, err = json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logsString += fmt.Sprintf("withdrawLogs = `%s`\n", logs)

	// WithdrawTo
	data = testsuite.PackWithdrawTo(amount, receiverAddressEth)
	tx, err = s.SendEthTransactionToSequencerInterfaceContract(data)
	s.Require().NoError(err)
	logs, err = json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logsString += fmt.Sprintf("withdrawToLogs = `%s`\n", logs)

	// Transfer
	data = testsuite.PackTransfer(receiverAddressEth, amount)
	tx, err = s.SendEthTransactionToSequencerInterfaceContract(data)
	s.Require().NoError(err)
	logs, err = json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logsString += fmt.Sprintf("transferLogs = `%s`\n", logs)

	// Vote
	voteProposalId := uint64(1)
	voteOption := uint32(govtypesv1.VoteOption_VOTE_OPTION_YES)
	voteMetadata := "metadata"
	data = testsuite.PackVote(voteProposalId, voteOption, voteMetadata)
	tx, err = s.SendEthTransactionToSequencerInterfaceContract(data)
	s.Require().NoError(err)
	logs, err = json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logsString += fmt.Sprintf("voteLogs = `%s`\n", logs)

	// SetRewardRecipient
	data = testsuite.PackSetRewardRecipient(receiverAddressEth)
	tx, err = s.SendEthTransactionToSequencerInterfaceContract(data)
	s.Require().NoError(err)
	logs, err = json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logsString += fmt.Sprintf("setRewardRecipientLogs = `%s`\n", logs)

	// Authorize
	amountCoin := sdk.NewCoin(testsuite.BridgeDenom, amountSDK)
	amountCoins := sdk.NewCoins(amountCoin)
	msgSendBz := s.E2ETestSuite.GenerateMsgSendBz(senderAddress, receiverAddress, amountCoins)
	authorizeData := testsuite.PackAuthorize(msgSendBz)
	tx, err = s.SendEthTransactionToSequencerInterfaceContract(authorizeData)
	s.Require().NoError(err)
	logs, err = json.Marshal(tx.Logs)
	s.Require().NoError(err)
	logsString += fmt.Sprintf("authorizeLogs = `%s`\n", logs)

	// Print the fixtures
	fmt.Println(fmt.Sprintf("amount = %d", amount.Uint64()))
	fmt.Println(fmt.Sprintf("bridgeDenom = \"%s\"", testsuite.BridgeDenom))
	fmt.Println(fmt.Sprintf("vestingDurationSeconds = \"%d\"", int64(vestingDuration.Seconds())))
	fmt.Println(fmt.Sprintf("voteProposalId = uint64(%d)", voteProposalId))
	fmt.Println(fmt.Sprintf("voteOption = int32(%d)", voteOption))
	fmt.Println(fmt.Sprintf("voteMetadata = \"%s\"", voteMetadata))
	fmt.Println(fmt.Sprintf("senderAddress = \"%s\"", senderAddress))
	fmt.Println(fmt.Sprintf("receiverAddress = \"%s\"", receiverAddress))
	fmt.Println(fmt.Sprintf("validator1Address = \"%s\"", validator1AddressEth))
	fmt.Println(fmt.Sprintf("validator2Address = \"%s\"", validator2AddressEth))
	fmt.Println(fmt.Sprintf("sequencerProxyContractAddress = \"%s\"", testsuite.SequencerProxyContractAddressStr))
	fmt.Println(logsString)
}
