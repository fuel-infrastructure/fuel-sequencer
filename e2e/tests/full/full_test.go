package full_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
	bridgemoduletypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// Ethereum RPC call
type RPCRequest struct {
	JsonRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

type EthLogsParams struct {
	FromBlock string `json:"fromBlock"`
	ToBlock   string `json:"toBlock"`
	Address   string `json:"address"`
}

func (s *FullTestSuite) TestWithdrawalWithCentralisedSolution_WithdrawalFromEthereum() {
	var fixtureEthereumDepositBlock string
	var fixtureEthereumAuthorizeBlock string
	var fixtureEthereumBridgeCommitmentBlock string
	var fixtureEthereumWithdrawalBlock string
	var fixtureEthereumSupplyDeltaBlock string
	var fixtureEthereumDepositWithLockupBlock string
	var fixtureEthereumDepositWithRecipientBlock string

	var fixtureSequencerWithdrawalBlock int64
	var fixtureSequencerMultipleWithdrawalsBlock int64
	var fixtureSequencerWithdrawalSupplyDeltaBlock int64

	var fixtureSequencerProofHeight int64
	var fixtureSequencerTxIndex int64
	var fixtureSequencerStartBlock uint64
	var fixtureSequencerEndBlock uint64

	var fixtureSequencerSupplyProofHeight int64
	var fixtureSequencerSupplyTxIndex int64
	var fixtureSequencerSupplyStartBlock uint64
	var fixtureSequencerSupplyEndBlock uint64

	s.Run("Submit a deposit to the Sequencer so that the Ethereum contract escrows the tokens", func() {
		senderAddress := s.EthKeys[0].AddressHex
		ownedReceiverAddressSeq := s.EthKeys[0].AddressSeq

		// --------------------------------------- Account owned by sender

		// Make sure that the balance of the receiver is as expected.
		expectedInitBalance := sdk.NewInt64Coin(testsuite.BridgeDenom, 0)
		balance, err := s.QueryAllBalances(s.Ctx(), ownedReceiverAddressSeq, nil)
		s.Require().NoError(err)
		s.Require().Equal(expectedInitBalance.Amount, balance.Balances.AmountOf(testsuite.BridgeDenom))

		// Deposit!
		sendAmount := big.NewInt(200)
		receiptDeposit := s.DepositTokenToSequencer(sendAmount)

		// Match the expected balance for the receiver on the Sequencer
		amountCoin := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(sendAmount))
		s.PollForBalance(s.Ctx(), 10, ownedReceiverAddressSeq, amountCoin)

		ethOwnedBaseAcc, err := s.QueryEthOwnedBaseAccount(s.Ctx(), ownedReceiverAddressSeq)
		s.Require().NoError(err)
		s.Require().Equal(senderAddress, ethOwnedBaseAcc.AccountOwner)

		fixtureEthereumDepositBlock = hexutil.EncodeUint64(receiptDeposit.BlockNumber.Uint64())
		fmt.Println(fmt.Sprintf("FIXTURE ETHEREUM: deposit on block %s", fixtureEthereumDepositBlock))

		// --------------------------------------- Authorize

		validator1Address := s.SeqKeys[0].ValAddressSeq
		delegatorAddress := s.EthKeys[0].AddressHex

		// Make sure that there is no pre-existing delegation between the delegator and validator1.
		delegationRaw, err := s.QueryDelegationRaw(s.Ctx(), delegatorAddress, validator1Address)
		s.Require().Nil(delegationRaw)
		s.Require().ErrorContains(
			err,
			fmt.Sprintf("delegation with delegator %s not found for validator %s", delegatorAddress, validator1Address),
		)

		// Generate Authorize event wrapping a MsgDelegate to validator1.
		delegateAmount, ok := sdkmath.NewIntFromString("100")
		s.Require().True(ok)
		delegateCoin := sdk.NewCoin(testsuite.BridgeDenom, delegateAmount)
		msgDelegateBz := s.E2ETestSuite.GenerateMsgDelegateBz(delegatorAddress, validator1Address, delegateCoin)
		authorizeData := testsuite.PackAuthorize(msgDelegateBz)
		receiptAuthorize, err := s.SendEthTransactionToSequencerInterfaceContract(authorizeData)
		s.Require().NoError(err)

		// Confirm that the delegation went through and is as expected.
		s.PollForDelegationBalance(s.Ctx(), 10, delegatorAddress, validator1Address, delegateCoin)

		fixtureEthereumAuthorizeBlock = hexutil.EncodeUint64(receiptAuthorize.BlockNumber.Uint64())
		fmt.Println(fmt.Sprintf("FIXTURE ETHEREUM: authorize on block %s", fixtureEthereumAuthorizeBlock))
	})

	s.Run("Submit a withdrawal from Ethereum and make sure it can be actioned on Ethereum", func() {

		withdrawerAddress := s.EthKeys[0].AddressHex
		withdrawCoin := sdk.NewInt64Coin(testsuite.BridgeDenom, 100)

		// --------------------------------------- Fund Ethereum owned account that will withdraw

		msgSend := &banktypes.MsgSend{
			FromAddress: s.SeqKeys[0].AddressSeq,
			ToAddress:   withdrawerAddress,
			Amount:      sdk.NewCoins(withdrawCoin),
		}
		res, err := s.SubmitMsgs(msgSend)
		s.Require().NoError(err)
		s.Require().Zero(res.Code)

		// --------------------------------------- User withdraws on the Sequencer via Ethereum

		sequencerHeightBefore, err := s.GetFuelSequencerHeight(s.Ctx())
		s.Require().NoError(err)

		// Generate Authorize event wrapping a MsgWithdrawToEthereum.
		msgWithdrawToEthereumBz := s.E2ETestSuite.GenerateMsgWithdrawToEthereumBz(
			withdrawerAddress, withdrawerAddress, withdrawCoin,
		)
		authorizeData := testsuite.PackAuthorize(msgWithdrawToEthereumBz)
		txReceipt, err := s.SendEthTransactionToSequencerInterfaceContract(authorizeData)
		s.Require().NoError(err)

		// The LastResultsHash is generated at the block right after the withdrawal
		s.PollForLastEthereumBlockSynced(s.Ctx(), 20, txReceipt.BlockNumber.Uint64())
		sequencerHeightAfter, err := s.GetFuelSequencerHeight(s.Ctx())
		s.Require().NoError(err)

		// Get withdrawal event type
		withdrawalEvent, err := sdk.TypedEventToEvent(&bridgemoduletypes.EventWithdrawToEthereumReported{})
		s.Require().NoError(err)
		withdrawalEventType := withdrawalEvent.Type

		// Find the actual LastResultsHash height
		var found bool
		var lastResultsHashHeight int64
		for i := sequencerHeightBefore; i <= sequencerHeightAfter; i++ {
			_, found = s.SearchForEventInBlockResults(s.Ctx(), withdrawalEventType, int64(i))
			if found {
				lastResultsHashHeight = int64(i + 1)
				break
			}
		}
		s.Require().True(found)
		fixtureSequencerWithdrawalBlock = lastResultsHashHeight - 1
		fmt.Println(fmt.Sprintf("FIXTURE SEQUENCER: withdrawal on block %d", fixtureSequencerWithdrawalBlock))

		// We wait two Sequencer blocks to ensure that we can capture the last results hash in the Bridge Commitment.
		s.Require().NoError(s.WaitForSequencerBlocks(s.Ctx(), 2, 10*time.Second))

		// Submit bridge commitment to FuelStreamX contract
		startBlock := uint64(1)
		endBlock := uint64(lastResultsHashHeight + 1)
		targetHeaderHash, bridgeCommitmentHash := s.GetDataForUpdateCommitHeaderRange(s.Ctx(), startBlock, endBlock)
		data := testsuite.PackUpdateCommitHeaderRangeMessage(endBlock, targetHeaderHash, bridgeCommitmentHash)
		receipt, err := s.SendEthTransactionToFuelStreamXContractAsGuardian(data)
		s.Require().NoError(err)

		// The two events are: HeadUpdate, BridgeCommitmentStored
		s.Require().Len(receipt.Logs, 2)

		// Extract log 1's topics and data
		s.Require().Len(receipt.Logs[1].Topics, 4)
		eventTopic0 := receipt.Logs[1].Topics[0].Hex()
		eventTopic1 := receipt.Logs[1].Topics[1].Hex()
		eventTopic2 := receipt.Logs[1].Topics[2].Hex()
		eventTopic3 := receipt.Logs[1].Topics[3].Hex()
		eventData := receipt.Logs[1].Data

		fixtureEthereumBridgeCommitmentBlock = hexutil.EncodeUint64(receipt.BlockNumber.Uint64())
		fmt.Println(fmt.Sprintf("FIXTURE ETHEREUM: bridge commitment %s", fixtureEthereumBridgeCommitmentBlock))

		fuelStreamxABI, err := abi.JSON(strings.NewReader(testsuite.FuelStreamXContractABI))
		s.Require().NoError(err)

		// Check BridgeCommitmentStored event

		expectedBridgeCommitment := s.QueryBridgeCommitment(s.Ctx(), startBlock, endBlock)

		actualStartBlock, err := strconv.ParseUint(eventTopic1[2:], 16, 64) // hex to uint64
		s.Require().NoError(err)
		actualTargetBlock, err := strconv.ParseUint(eventTopic2[2:], 16, 64) // hex to uint64
		s.Require().NoError(err)

		s.Require().Equal(testsuite.BridgeCommitmentStoredEventHash, eventTopic0)
		s.Require().EqualValues(startBlock, actualStartBlock)
		s.Require().EqualValues(endBlock, actualTargetBlock)
		s.Require().Equal(expectedBridgeCommitment.String(), strings.ToUpper(eventTopic3[2:]))

		var event testsuite.BridgeCommitmentStoredEvent
		err = fuelStreamxABI.UnpackIntoInterface(&event, testsuite.BridgeCommitmentStoredEventName, eventData)
		s.Require().NoError(err)

		expectedProofNonce := 1
		s.Require().EqualValues(expectedProofNonce, event.ProofNonce.Uint64())

		// --------------------------------------- Make a withdrawal on Ethereum

		// Get BridgeCommitment inclusion proof
		// - The 'last result hash' incorporating the withdrawal result is at h+1.
		// - The withdrawal is assumed to be the second transaction in the block, following the MsgIndex.
		txIndex := int64(1) // second tx
		bcLeaf, bcLeafProof, txResultMarshalled, txResultProof := s.GetDataForBridgeCommitmentInclusionProof(
			s.Ctx(), lastResultsHashHeight, txIndex, startBlock, endBlock,
		)

		fixtureSequencerProofHeight = lastResultsHashHeight
		fixtureSequencerTxIndex = txIndex
		fixtureSequencerStartBlock = startBlock
		fixtureSequencerEndBlock = endBlock
		fmt.Println(fmt.Sprintf(
			"FIXTURE SEQUENCER: proof height %d & txIndex %d & startBlock %d & endBlock %d",
			fixtureSequencerProofHeight, fixtureSequencerTxIndex, fixtureSequencerStartBlock, fixtureSequencerEndBlock),
		)

		// Submit transaction to Ethereum to process the withdrawal.
		data = testsuite.PackProcessSequencerWithdrawalMessage(
			event.ProofNonce, bcLeaf, bcLeafProof, txResultMarshalled, txResultProof,
		)
		receiptWithdrawal, err := s.SendEthTransactionToFuelStreamXContractAsUser(data)
		s.Require().NoError(err)

		fixtureEthereumWithdrawalBlock = hexutil.EncodeUint64(receiptWithdrawal.BlockNumber.Uint64())
		fmt.Println(fmt.Sprintf("FIXTURE ETHEREUM: withdrawal on block %s", fixtureEthereumWithdrawalBlock))

		// ------------------------------------- Supply delta execution

		lastResultsHashHeightSupplyDelta := int64(6)
		bcLeaf, bcLeafProof, txResultMarshalled, txResultProof = s.GetDataForBridgeCommitmentInclusionProof(
			s.Ctx(), lastResultsHashHeightSupplyDelta, txIndex, startBlock, endBlock,
		)

		fixtureSequencerSupplyProofHeight = lastResultsHashHeightSupplyDelta
		fixtureSequencerSupplyTxIndex = txIndex
		fixtureSequencerSupplyStartBlock = startBlock
		fixtureSequencerSupplyEndBlock = endBlock
		fmt.Println(fmt.Sprintf(
			"FIXTURE SEQUENCER: proof height %d & txIndex %d & startBlock %d & endBlock %d",
			fixtureSequencerSupplyProofHeight,
			fixtureSequencerSupplyTxIndex,
			fixtureSequencerSupplyStartBlock,
			fixtureSequencerSupplyEndBlock,
		))

		dataSupplyDelta := testsuite.PackProcessSequencerSupplyDeltaMessage(
			event.ProofNonce, bcLeaf, bcLeafProof, txResultMarshalled, txResultProof,
		)

		receiptSupplyDelta, err := s.SendEthTransactionToFuelStreamXContractAsUser(dataSupplyDelta)
		s.Require().NoError(err)

		fixtureEthereumSupplyDeltaBlock = hexutil.EncodeUint64(receiptSupplyDelta.BlockNumber.Uint64())
		fmt.Println(fmt.Sprintf("FIXTURE ETHEREUM: supply delta on block %s", fixtureEthereumSupplyDeltaBlock))

		// -------------------------------------- Multiple withdrawals in same transaction

		withdrawMsg1 := bridgemoduletypes.NewMsgWithdrawToEthereum(
			s.SeqKeys[0].AddressSeq,
			"0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
			sdk.NewInt64Coin(testsuite.BridgeDenom, 123),
		)
		withdrawMsg2 := bridgemoduletypes.NewMsgWithdrawToEthereum(
			s.SeqKeys[0].AddressSeq,
			"0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
			sdk.NewInt64Coin(testsuite.BridgeDenom, 400),
		)
		withdrawalResponse, err := s.SubmitMsgs(withdrawMsg1, withdrawMsg2)
		s.Require().NoError(err)
		s.Require().Zero(withdrawalResponse.Code)

		fixtureSequencerMultipleWithdrawalsBlock = withdrawalResponse.Height
		fmt.Println(fmt.Sprintf("FIXTURE SEQUENCER: mutliple withdrawals %d", fixtureSequencerMultipleWithdrawalsBlock))

		// -------------------------------------- Withdrawal and Supply delta in same block

		for {
			withdrawMsgInfinite := bridgemoduletypes.NewMsgWithdrawToEthereum(
				s.SeqKeys[0].AddressSeq,
				"0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
				sdk.NewInt64Coin(testsuite.BridgeDenom, 100),
			)
			withdrawalResponse, err := s.SubmitMsgs(withdrawMsgInfinite)
			s.Require().NoError(err)
			s.Require().Zero(withdrawalResponse.Code)

			if withdrawalResponse.Height%5 == 0 {
				fixtureSequencerWithdrawalSupplyDeltaBlock = withdrawalResponse.Height
				fmt.Println(fmt.Sprintf("FIXTURE SEQUENCER: withdrawal + supply delta %d", fixtureSequencerWithdrawalSupplyDeltaBlock))
				break
			}
		}

		// -------------------------------------- Deposit with lockup
		sendAmount := big.NewInt(246)
		receiptDepositWithLockup := s.DepositTokenToSequencerFromMigrationNoDelegation(sendAmount)

		fixtureEthereumDepositWithLockupBlock = hexutil.EncodeUint64(receiptDepositWithLockup.BlockNumber.Uint64())
		fmt.Println(fmt.Sprintf("FIXTURE ETHEREUM: deposit with lockup on block %s", fixtureEthereumDepositWithLockupBlock))

		// -------------------------------------- Deposit with recipient

		recipient := common.HexToAddress("0x4838B106FCe9647Bdf1E7877BF73cE8B0BAD5f97")
		receiptDepositWithRecipient := s.DepositForToSequencer(sendAmount, recipient)

		fixtureEthereumDepositWithRecipientBlock = hexutil.EncodeUint64(receiptDepositWithRecipient.BlockNumber.Uint64())
		fmt.Println(fmt.Sprintf("FIXTURE ETHEREUM: deposit with different recipient on block %s", fixtureEthereumDepositWithRecipientBlock))

		// ------- Sequencer

		s.apiCallHelperTendermint("http://localhost:26657/", fmt.Sprintf("block?height=%d", 5)) // Supply delta
		s.apiCallHelperTendermint("http://localhost:26657/", fmt.Sprintf("block?height=%d", fixtureSequencerWithdrawalBlock))
		s.apiCallHelperTendermint("http://localhost:26657/", fmt.Sprintf("block?height=%d", fixtureSequencerMultipleWithdrawalsBlock))
		s.apiCallHelperTendermint("http://localhost:26657/", fmt.Sprintf("block?height=%d", fixtureSequencerWithdrawalSupplyDeltaBlock))

		s.apiCallHelperTendermint("http://localhost:26657/", fmt.Sprintf("block_results?height=%d", 5)) // Supply delta
		s.apiCallHelperTendermint("http://localhost:26657/", fmt.Sprintf("block_results?height=%d", fixtureSequencerWithdrawalBlock))
		s.apiCallHelperTendermint("http://localhost:26657/", fmt.Sprintf("block_results?height=%d", fixtureSequencerMultipleWithdrawalsBlock))
		s.apiCallHelperTendermint("http://localhost:26657/", fmt.Sprintf("block_results?height=%d", fixtureSequencerWithdrawalSupplyDeltaBlock))

		s.apiCallHelperTendermint("http://localhost:1317/fuelsequencer/commitments/v1/", fmt.Sprintf("bridge_commitment_inclusion_proof?height=%d&tx_index=%d&start=%d&end=%d",
			fixtureSequencerProofHeight,
			fixtureSequencerTxIndex,
			fixtureSequencerStartBlock,
			fixtureSequencerEndBlock,
		))
		s.apiCallHelperTendermint("http://localhost:1317/fuelsequencer/commitments/v1/", fmt.Sprintf("bridge_commitment_inclusion_proof?height=%d&tx_index=%d&start=%d&end=%d",
			fixtureSequencerSupplyProofHeight,
			fixtureSequencerSupplyTxIndex,
			fixtureSequencerSupplyStartBlock,
			fixtureSequencerSupplyEndBlock,
		))

		// ------- Ethereum

		blocks := []string{
			fixtureEthereumDepositBlock,
			fixtureEthereumAuthorizeBlock,
			fixtureEthereumBridgeCommitmentBlock,
			fixtureEthereumWithdrawalBlock,
			fixtureEthereumSupplyDeltaBlock,
			fixtureEthereumDepositWithLockupBlock,
			fixtureEthereumDepositWithRecipientBlock,
		}

		for i, block := range blocks {
			s.apiCallHelperEthereum("http://localhost:8545", RPCRequest{
				JsonRPC: "2.0",
				Method:  "eth_getBlockByNumber",
				Params:  []interface{}{block, false},
				ID:      i + 1,
			}, fmt.Sprintf("eth_getBlockByNumber?height=%s", block))
		}

		s.apiCallHelperEthereum("http://localhost:8545", RPCRequest{
			JsonRPC: "2.0",
			Method:  "eth_getLogs",
			Params: []interface{}{EthLogsParams{
				FromBlock: fixtureEthereumDepositBlock,
				ToBlock:   fixtureEthereumDepositBlock,
				Address:   s.GetSequencerProxyAddress().String(),
			}},
			ID: 1,
		}, fmt.Sprintf("eth_getLogs?height=%s", fixtureEthereumDepositBlock))

		s.apiCallHelperEthereum("http://localhost:8545", RPCRequest{
			JsonRPC: "2.0",
			Method:  "eth_getLogs",
			Params: []interface{}{EthLogsParams{
				FromBlock: fixtureEthereumAuthorizeBlock,
				ToBlock:   fixtureEthereumAuthorizeBlock,
				Address:   s.GetSequencerProxyAddress().String(),
			}},
			ID: 1,
		}, fmt.Sprintf("eth_getLogs?height=%s", fixtureEthereumAuthorizeBlock))

		s.apiCallHelperEthereum("http://localhost:8545", RPCRequest{
			JsonRPC: "2.0",
			Method:  "eth_getLogs",
			Params: []interface{}{EthLogsParams{
				FromBlock: fixtureEthereumBridgeCommitmentBlock,
				ToBlock:   fixtureEthereumBridgeCommitmentBlock,
				Address:   s.GetFuelStreamXAddress().String(),
			}},
			ID: 1,
		}, fmt.Sprintf("eth_getLogs?height=%s", fixtureEthereumBridgeCommitmentBlock))

		s.apiCallHelperEthereum("http://localhost:8545", RPCRequest{
			JsonRPC: "2.0",
			Method:  "eth_getLogs",
			Params: []interface{}{EthLogsParams{
				FromBlock: fixtureEthereumWithdrawalBlock,
				ToBlock:   fixtureEthereumWithdrawalBlock,
				Address:   s.GetFuelStreamXAddress().String(),
			}},
			ID: 1,
		}, fmt.Sprintf("eth_getLogs?height=%s", fixtureEthereumWithdrawalBlock))

		s.apiCallHelperEthereum("http://localhost:8545", RPCRequest{
			JsonRPC: "2.0",
			Method:  "eth_getLogs",
			Params: []interface{}{EthLogsParams{
				FromBlock: fixtureEthereumSupplyDeltaBlock,
				ToBlock:   fixtureEthereumSupplyDeltaBlock,
				Address:   s.GetFuelStreamXAddress().String(),
			}},
			ID: 1,
		}, fmt.Sprintf("eth_getLogs?height=%s", fixtureEthereumSupplyDeltaBlock))

		s.apiCallHelperEthereum("http://localhost:8545", RPCRequest{
			JsonRPC: "2.0",
			Method:  "eth_getLogs",
			Params: []interface{}{EthLogsParams{
				FromBlock: fixtureEthereumDepositWithLockupBlock,
				ToBlock:   fixtureEthereumDepositWithLockupBlock,
				Address:   s.GetSequencerProxyAddress().String(),
			}},
			ID: 1,
		}, fmt.Sprintf("eth_getLogs?height=%s", fixtureEthereumDepositWithLockupBlock))

		s.apiCallHelperEthereum("http://localhost:8545", RPCRequest{
			JsonRPC: "2.0",
			Method:  "eth_getLogs",
			Params: []interface{}{EthLogsParams{
				FromBlock: fixtureEthereumDepositWithRecipientBlock,
				ToBlock:   fixtureEthereumDepositWithRecipientBlock,
				Address:   s.GetSequencerProxyAddress().String(),
			}},
			ID: 1,
		}, fmt.Sprintf("eth_getLogs?height=%s", fixtureEthereumDepositWithRecipientBlock))
	})
}

func (s *FullTestSuite) apiCallHelperTendermint(url string, query string) {
	response, err := http.Get(fmt.Sprintf("%s%s", url, query))
	s.Require().NoError(err)
	s.Require().True(response.StatusCode == http.StatusOK)
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	s.Require().NoError(err)

	// Pretty print the JSON for better readability
	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, body, "", "  ")
	s.Require().NoError(err)

	err = ioutil.WriteFile(fmt.Sprintf("fixtures/sequencer/%s.json", query), prettyJSON.Bytes(), 0644)
	s.Require().NoError(err)
}

func (s *FullTestSuite) apiCallHelperEthereum(url string, query RPCRequest, fileName string) {
	jsonBody, err := json.Marshal(query)
	s.Require().NoError(err)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	s.Require().NoError(err)

	// Pretty print the JSON for better readability
	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, body, "", "  ")
	s.Require().NoError(err)

	err = ioutil.WriteFile(fmt.Sprintf("fixtures/ethereum/%s.json", fileName), prettyJSON.Bytes(), 0644)
	s.Require().NoError(err)
}
