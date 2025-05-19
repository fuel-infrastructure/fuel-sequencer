package upgrades_test

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/fuel-infrastructure/fuel-sequencer/app/upgrades/bond_module"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
	bondtypes "github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

const (
	bondHaltHeightDelta        = uint64(25) // will propose upgrade this many blocks in the future; must be > voting period
	bondBlocksAfterUpgrade     = uint64(10) // will wait for this many blocks after the upgrade
	bondModuleFromImageVersion = "7d60123"  // this image needs to exist for this test to run
	bondModuleToImageVersion   = "d1fc0d4"  // this will be updated as work progresses

)

var (
	yieldRecipient = func(s *BondModuleUpgradeTestSuite) string {
		return s.SeqKeys[1].AddressSeq // alice's address
	}
	yieldAmount     = sdkmath.NewInt(1000000)
	sharedYieldTime *time.Time
	yieldTime       = func() *time.Time {
		if sharedYieldTime == nil {
			t := time.Now().Add(1 * time.Minute)
			sharedYieldTime = &t
		}
		return sharedYieldTime
	}
)

type BondModuleUpgradeTestSuite struct {
	testsuite.E2ETestSuite
}

func TestBondModuleUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(BondModuleUpgradeTestSuite))
}

func (s *BondModuleUpgradeTestSuite) SetupTest() {
	s.FuelSequencerDockerImageTag = bondModuleFromImageVersion
	s.E2ETestSuite.SetupTest()
}

func (s *BondModuleUpgradeTestSuite) TestBondModuleUpgrade() {
	s.Run("Perform the upgrade", func() {
		height, err := s.GetFuelSequencerHeight(s.Ctx())
		s.Require().NoError(err, "error fetching height before submit upgrade proposal")
		s.Logger().Info("Current height before upgrade proposal", zap.Int64("height", int64(height)))

		haltHeight := height + bondHaltHeightDelta
		s.Logger().Info("Submitting software upgrade proposal",
			zap.Uint64("halt_height", haltHeight),
			zap.String("upgrade_name", bond_module.UpgradeName))
		msgUpgrade := &upgradetypes.MsgSoftwareUpgrade{
			Authority: s.GetGovernanceAddress(),
			Plan: upgradetypes.Plan{
				Name:   bond_module.UpgradeName,
				Height: int64(haltHeight),
				Info:   "<dummy-info>",
			},
		}
		s.ExecuteGovProposal(msgUpgrade)

		_, err = s.GetFuelSequencerHeight(s.Ctx())
		s.Require().NoError(err, "error fetching height before upgrade")

		// Wait until just before the upgrade
		err = s.WaitUntilSequencerBlock(s.Ctx(), int(haltHeight-1), time.Second*20)
		s.Require().NoError(err)

		// Hold Ethereum so the Sequencer doesn't get too out-of-sync
		s.PauseEthereum()

		// Ensure the nodes have reached the halt height
		time.Sleep(time.Second * 5)

		// Bring down nodes to prepare for upgrade.
		s.Logger().Info("Stopping all sequencer nodes...")
		s.StopAllSequencerNodes()
		s.Logger().Info("Removing all sequencer nodes...")
		s.RemoveAllSequencerNodes()

		// Resume Ethereum since we're about to resume the Sequencer
		s.UnpauseEthereum()

		// Upgrade version on all nodes and start them back up.
		s.Logger().Info("Starting nodes back up...")
		s.FuelSequencerDockerImageTag = bondModuleToImageVersion
		s.RunSequencerValidators()

		err = s.WaitForSequencerBlocks(s.Ctx(), int(bondBlocksAfterUpgrade), time.Second*20)
		s.Require().NoError(err, "chain did not produce blocks after upgrade")
	})

	// Check that bond module params and state are now queryable
	// Upgrade handler should modify params on upgrade
	s.Run("Bond Params and State start as default, are updatable and queryable", func() {
		// Get the bond module params
		bondParams := s.QueryBondParams(s.Ctx())
		s.Require().NotNil(bondParams)
		s.Logger().Info("Initial bond params",
			zap.String("yield_recipient", bondParams.YieldRecipient),
			zap.Any("yield_time", bondParams.YieldTime),
			zap.String("yield_amount", bondParams.YieldAmount.String()))

		// Check that the yield recipient is set to default value
		s.Require().Equal("", bondParams.YieldRecipient)

		// Check that the yield time is set to default value
		s.Require().Nil(bondParams.YieldTime)

		// Check that the yield amount is set to default value
		s.Require().True(bondParams.YieldAmount.IsZero())

		// Get the bond module state
		bondState := s.QueryBondState(s.Ctx())
		s.Require().NotNil(bondState)
		s.Logger().Info("Initial bond state",
			zap.Int64("yield_mint_height", bondState.YieldMintHeight))

		// Check that the yield mint height is set to default value
		s.Require().Equal(int64(0), bondState.YieldMintHeight)

		// Propose a new bond params via an expedited proposal
		bondParams.YieldRecipient = yieldRecipient(s)
		bondParams.YieldTime = yieldTime()
		bondParams.YieldAmount = yieldAmount

		s.Logger().Info("Proposing new bond params",
			zap.String("new_yield_recipient", bondParams.YieldRecipient),
			zap.Time("new_yield_time", *bondParams.YieldTime),
			zap.String("new_yield_amount", bondParams.YieldAmount.String()))

		msgUpdateBond := &bondtypes.MsgUpdateParams{
			Authority: s.GetGovernanceAddress(),
			Params:    *bondParams,
		}
		s.ExecuteExpeditedGovProposal(msgUpdateBond)

		// Wait for 1 block to pass for the BeginBlocker to run
		s.WaitForSequencerBlocks(s.Ctx(), 1, time.Second*10)

		// Verify params were updated
		updatedBondParams := s.QueryBondParams(s.Ctx())
		s.Logger().Info("Updated bond params",
			zap.String("yield_recipient", updatedBondParams.YieldRecipient),
			zap.Any("yield_time", updatedBondParams.YieldTime),
			zap.String("yield_amount", updatedBondParams.YieldAmount.String()))

		s.Require().Equal(yieldRecipient(s), updatedBondParams.YieldRecipient)
		s.Require().Equal(yieldTime().Unix(), updatedBondParams.YieldTime.Unix())
		s.Require().Equal(yieldAmount, updatedBondParams.YieldAmount)

		// Verify state is still queryable and unchanged
		updatedBondState := s.QueryBondState(s.Ctx())
		s.Require().NotNil(updatedBondState)
		s.Logger().Info("Updated bond state",
			zap.Int64("yield_mint_height", updatedBondState.YieldMintHeight))
		s.Require().Equal(int64(0), updatedBondState.YieldMintHeight, "state should remain unchanged after params update")
	})

	// With the updated params, confirm that the bond module:
	// 1. Mints the requested amount of tokens
	// 2. The total supply is increased by the minted amount
	// 3. Yields the requested amount of tokens to the yield recipient
	// 4. The State of the bond module is updated to show the block height of the yield time
	s.Run("Bond module yields as intended", func() {
		denom := s.QueryBridgeParams(s.Ctx()).BridgeDenom

		// Get initial state
		initialSupply, err := s.QuerySupply(s.Ctx(), denom)
		s.Require().NoError(err)
		recipient := yieldRecipient(s)
		initialRecipientBalance, err := s.QueryBalance(s.Ctx(), recipient, denom)
		s.Require().NoError(err)

		s.Logger().Info("Initial state before yield",
			zap.String("initial_supply", initialSupply.String()),
			zap.String("initial_recipient_balance", initialRecipientBalance.Balance.Amount.String()))

		// Verify bond params from previous test case
		bondParams := s.QueryBondParams(s.Ctx())
		s.Logger().Info("Current bond params before yield",
			zap.String("yield_recipient", bondParams.YieldRecipient),
			zap.Any("yield_time", bondParams.YieldTime),
			zap.String("yield_amount", bondParams.YieldAmount.String()))

		// Verify params are set from previous test case
		s.Require().Equal(yieldRecipient(s), bondParams.YieldRecipient)
		s.Require().Equal(yieldTime().Unix(), bondParams.YieldTime.Unix())
		s.Require().Equal(yieldAmount, bondParams.YieldAmount)

		// Check if yield has already been minted
		bondState := s.QueryBondState(s.Ctx())
		s.Logger().Info("Current bond state",
			zap.Int64("yield_mint_height", bondState.YieldMintHeight))

		if bondState.YieldMintHeight > 0 {
			s.Logger().Info("Yield has already been minted at height",
				zap.Int64("height", bondState.YieldMintHeight))
		} else {
			// Wait until we reach or pass the yield time and yield occurs
			for {
				currentTime := time.Now()
				bondState = s.QueryBondState(s.Ctx())

				// If minting has occurred, verify it happened at the right time
				if bondState.YieldMintHeight > 0 {
					// Get the block time for the mint height
					block, err := s.GetBlockByHeight(s.Ctx(), bondState.YieldMintHeight)
					s.Require().NoError(err, "failed to get block at mint height")
					mintTime := block.Header.Time

					// Log detailed timing information
					s.Logger().Info("Yield minting occurred",
						zap.Int64("height", bondState.YieldMintHeight),
						zap.Time("requested_yield_time", *bondParams.YieldTime),
						zap.Time("actual_mint_time", mintTime),
						zap.String("time_difference", mintTime.Sub(*bondParams.YieldTime).String()),
						zap.Int64("current_block_height", bondState.YieldMintHeight),
						zap.Int64("yield_mint_height", bondState.YieldMintHeight),
					)

					// Verify that minting occurred at or after the intended yield time
					s.Require().True(bondParams.YieldTime.Before(mintTime) || bondParams.YieldTime.Equal(mintTime),
						"yield minting occurred before intended yield time")
					break
				}

				// If we haven't reached yield time yet, wait and continue
				if !currentTime.After(*bondParams.YieldTime) {
					s.Logger().Info("Current time before yield time",
						zap.Time("current_time", currentTime),
						zap.Time("yield_time", *bondParams.YieldTime))
					time.Sleep(time.Second)
					continue
				}

				// We've reached yield time, wait for two blocks to ensure minting occurs in the first block after yield time
				err = s.WaitForSequencerBlocks(s.Ctx(), 2, time.Second*10)
				s.Require().NoError(err, "failed to wait for blocks after yield time")

				// After expected yield time, final check after the upcoming block to assert that minting occurred
				bondState = s.QueryBondState(s.Ctx())
				if bondState.YieldMintHeight > 0 {
					// Get the block time for the mint height
					block, err := s.GetBlockByHeight(s.Ctx(), bondState.YieldMintHeight)
					s.Require().NoError(err, "failed to get block at mint height")
					mintTime := block.Header.Time

					// Log detailed timing information
					s.Logger().Info("Yield minting occurred after waiting",
						zap.Int64("height", bondState.YieldMintHeight),
						zap.Time("requested_yield_time", *bondParams.YieldTime),
						zap.Time("actual_mint_time", mintTime),
						zap.String("time_difference", mintTime.Sub(*bondParams.YieldTime).String()),
						zap.Int64("current_block_height", bondState.YieldMintHeight),
						zap.Int64("yield_mint_height", bondState.YieldMintHeight),
					)

					// Verify that minting occurred at or after the intended yield time
					s.Require().True(
						bondParams.YieldTime.Before(mintTime) || bondParams.YieldTime.Equal(mintTime),
						"yield minting occurred before intended yield time")
					break
				}

				// If we've waited a block after yield time and still no minting, fail
				s.Require().Fail("Yield minting did not occur within one block after yield time")
				return
			}
		}

		// Verify total supply increased by yield amount and log all supply data
		finalSupply, err := s.QuerySupply(s.Ctx(), denom)
		s.Require().NoError(err)
		s.Logger().Info("Token supply data w.r.t mint",
			zap.String("initial_supply", initialSupply.String()),
			zap.String("yield_amount", bondParams.YieldAmount.String()),
			zap.String("expected_final_supply", initialSupply.Add(bondParams.YieldAmount).String()),
			zap.String("actual_final_supply", finalSupply.String()),
			zap.String("supply_difference", finalSupply.Sub(initialSupply).String()),
		)

		// Verify recipient balance increased by yield amount and log all balance data
		finalRecipientBalance, err := s.QueryBalance(s.Ctx(), recipient, denom)
		s.Require().NoError(err)
		s.Logger().Info("Recipient balance data w.r.t yield",
			zap.String("initial_recipient_balance", initialRecipientBalance.Balance.Amount.String()),
			zap.String("expected_final_balance", initialRecipientBalance.Balance.Amount.Add(bondParams.YieldAmount).String()),
			zap.String("actual_final_balance", finalRecipientBalance.Balance.Amount.String()),
			zap.String("balance_difference", finalRecipientBalance.Balance.Amount.Sub(initialRecipientBalance.Balance.Amount).String()),
		)

		// Now perform assertions
		s.Require().Equal(initialSupply.Add(bondParams.YieldAmount), finalSupply,
			"total supply should increase by yield amount")
		s.Require().Equal(initialRecipientBalance.Balance.Amount.Add(bondParams.YieldAmount), finalRecipientBalance.Balance.Amount,
			"recipient balance should increase by yield amount")
	})

	// With the upgraded instance, ensure basic operations are working as usual
	// This will be tested by executing a deposit, followed by a delegation, and then an undelegation.
	s.Run("Ensure deposit and delegate working as usual (regression check)", func() {
		// Get and verify initial account balance
		senderAddress := s.EthKeys[0].AddressHex           // The depositor on Ethereum
		ownedReceiverAddressSeq := s.EthKeys[0].AddressSeq // Deposit receiver; owned by the sender
		validatorAddressHex := s.SeqKeys[0].ValAddressHex  // Address of one of the validators

		s.Logger().Info("starting regression check with addresses",
			zap.String("sender_address", senderAddress),
			zap.String("receiver_address", ownedReceiverAddressSeq),
			zap.String("validator_address", validatorAddressHex))

		// Make sure that the balance of the receiver is as expected.
		expectedInitBalance := sdk.NewInt64Coin(testsuite.BridgeDenom, 0)
		balance, err := s.QueryAllBalances(s.Ctx(), ownedReceiverAddressSeq, nil)
		s.Require().NoError(err)
		s.Require().Equal(expectedInitBalance.Amount, balance.Balances.AmountOf(testsuite.BridgeDenom))
		s.Logger().Info("initial balance check passed",
			zap.String("expected_balance", expectedInitBalance.Amount.String()),
			zap.String("actual_balance", balance.Balances.AmountOf(testsuite.BridgeDenom).String()))

		// Get account delegation balance(s)
		delegatorAddress := s.EthKeys[0].AddressHex
		validatorAddress := s.SeqKeys[0].ValAddressSeq

		// Make sure that there is no pre-existing delegation between the delegator and validator
		delegationRaw, err := s.QueryDelegationRaw(s.Ctx(), delegatorAddress, validatorAddress)
		s.Require().Nil(delegationRaw)
		s.Require().ErrorContains(
			err,
			fmt.Sprintf("delegation with delegator %s not found for validator %s", delegatorAddress, validatorAddress),
		)
		s.Logger().Info("verified no pre-existing delegation",
			zap.String("delegator", delegatorAddress),
			zap.String("validator", validatorAddress))

		// Get and verify initial validator delegations
		_, err = s.QueryDelegationRaw(s.Ctx(), delegatorAddress, validatorAddress)
		s.Require().Error(err) // Should error since no delegation exists yet
		s.Logger().Info("confirmed no initial delegation exists", zap.String("error", err.Error()))

		// Perform and verify account deposit
		sendAmount := big.NewInt(200)
		s.Logger().Info("initiating deposit",
			zap.String("amount", sendAmount.String()),
			zap.String("from", senderAddress),
			zap.String("to", ownedReceiverAddressSeq))
		receipt := s.DepositTokenToSequencer(sendAmount)
		s.Require().NotNil(receipt, "deposit receipt should not be nil")
		s.Require().Equal(uint64(1), receipt.Status, "deposit transaction should be successful")
		s.Logger().Info("deposit transaction receipt",
			zap.String("tx_hash", receipt.TxHash.Hex()),
			zap.Uint64("block_number", receipt.BlockNumber.Uint64()),
			zap.Uint64("status", receipt.Status))

		// Match the expected balance for the receiver on the Sequencer
		amountCoin := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(sendAmount))
		s.PollForBalance(s.Ctx(), 10, ownedReceiverAddressSeq, amountCoin)
		s.Logger().Info("deposit confirmed",
			zap.String("expected_amount", amountCoin.Amount.String()),
			zap.String("receiver", ownedReceiverAddressSeq))

		// Verify account ownership
		ethOwnedBaseAcc, err := s.QueryEthOwnedBaseAccount(s.Ctx(), ownedReceiverAddressSeq)
		s.Require().NoError(err)
		s.Require().Equal(senderAddress, ethOwnedBaseAcc.AccountOwner)
		s.Logger().Info("account ownership verified",
			zap.String("account", ownedReceiverAddressSeq),
			zap.String("owner", ethOwnedBaseAcc.AccountOwner))

		// Perform and verify account delegation
		validatorAddressEth := common.HexToAddress(validatorAddressHex)
		s.Logger().Info("initiating delegation",
			zap.String("amount", sendAmount.String()),
			zap.String("delegator", delegatorAddress),
			zap.String("validator", validatorAddressHex))
		delegationReceipt := s.DelegateTokenToSequencer(sendAmount, validatorAddressEth)
		s.Require().NotNil(delegationReceipt, "delegation receipt should not be nil")
		s.Require().Equal(uint64(1), delegationReceipt.Status, "delegation transaction should be successful")
		s.Logger().Info("delegation transaction receipt",
			zap.String("tx_hash", delegationReceipt.TxHash.Hex()),
			zap.Uint64("block_number", delegationReceipt.BlockNumber.Uint64()),
			zap.Uint64("status", delegationReceipt.Status))

		// Verify delegation was successful
		s.PollForDelegationBalance(s.Ctx(), 10, delegatorAddress, validatorAddressHex, amountCoin)
		s.Logger().Info("delegation confirmed",
			zap.String("amount", amountCoin.Amount.String()),
			zap.String("delegator", delegatorAddress),
			zap.String("validator", validatorAddressHex))

		// Verify validator delegations were updated
		delegation, err := s.QueryDelegationRaw(s.Ctx(), delegatorAddress, validatorAddress)
		s.Require().NoError(err)
		s.Require().Equal(amountCoin.Amount, delegation.DelegationResponse.Balance.Amount)
		s.Logger().Info("delegation balance verified",
			zap.String("expected_amount", amountCoin.Amount.String()),
			zap.String("actual_amount", delegation.DelegationResponse.Balance.Amount.String()))

		// Perform and verify account undelegation
		// Override unbonding time so that undelegation goes through immediately
		stakingParams := s.QueryStakingParams(s.Ctx())
		originalUnbondingTime := stakingParams.UnbondingTime
		stakingParams.UnbondingTime = time.Second
		s.Logger().Info("updating staking params for quick unbonding",
			zap.Duration("original_unbonding_time", originalUnbondingTime),
			zap.Duration("new_unbonding_time", stakingParams.UnbondingTime))
		s.ExecuteGovProposal(&stakingtypes.MsgUpdateParams{
			Authority: s.GetGovernanceAddress(),
			Params:    *stakingParams,
		})

		// Ensure value updated
		updatedParams := s.QueryStakingParams(s.Ctx())
		s.Require().Equal(time.Second, updatedParams.UnbondingTime)
		s.Logger().Info("staking params updated successfully",
			zap.Duration("current_unbonding_time", updatedParams.UnbondingTime))

		// Undelegate the previously delegated amount
		s.Logger().Info("initiating undelegation",
			zap.String("amount", sendAmount.String()),
			zap.String("delegator", delegatorAddress),
			zap.String("validator", validatorAddressHex))
		unbondData := testsuite.PackUnbond(sendAmount, validatorAddressEth)
		undelegation, err := s.SendEthTransactionToSequencerInterfaceContract(unbondData)
		s.Require().NoError(err)
		s.Logger().Info("undelegation transaction submitted",
			zap.String("tx_hash", undelegation.TxHash.Hex()),
			zap.Uint64("block_number", undelegation.BlockNumber.Uint64()))

		s.PollForLastEthereumBlockSynced(s.Ctx(), 10, undelegation.BlockNumber.Uint64()) // wait until tx processed
		s.PollForNoDelegation(s.Ctx(), 0, delegatorAddress, validatorAddress)
		s.Logger().Info("undelegation confirmed on chain")

		// Wait for undelegation to go through (Note: unbonding time is very small)
		s.WaitForSequencerBlocks(s.Ctx(), 1, time.Second*10)

		// Verify validator delegations were cleared
		_, err = s.QueryDelegationRaw(s.Ctx(), delegatorAddress, validatorAddress)
		s.Require().Error(err) // Should error since delegation was removed
		s.Logger().Info("verified delegation was removed",
			zap.String("delegator", delegatorAddress),
			zap.String("validator", validatorAddress),
			zap.String("error for no delegation", err.Error()),
		)

		// Verify final balance
		finalBalance, err := s.QueryAllBalances(s.Ctx(), ownedReceiverAddressSeq, nil)
		s.Require().NoError(err)
		s.Logger().Info("final balance check",
			zap.String("address", ownedReceiverAddressSeq),
			zap.String("final_balance", finalBalance.Balances.AmountOf(testsuite.BridgeDenom).String()))
	})
}
