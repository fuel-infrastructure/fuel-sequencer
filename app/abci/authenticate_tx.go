package abci

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// authenticateEvent has the responsibility of ensuring that the necessary authentication is in place for the event to
// be executed, and that any messages resulting from the encoded transaction are valid. An error returned from this
// function can cause block production to stop, whereas no authentication just means the event should just be skipped.
//
// For deposit events, this function is a no-op since MsgDepositFromEthereum does not have a ValidateBasic. We also
// don't check if the depositor address is blocked. This is intentional, since we want to perform some logic in the
// message handler even if some values are invalid, ensuring that we always mint the deposited amount.
//
// Note: the error always takes priority over the value of the returned bool.
func (h *FuelSequencerProposalHandler) authenticateEvent(
	event *sidecartypes.Event,
	rawTxBytes []byte,
	params *bridgetypes.Params,
	blockedBech32Addresses map[string]bool,
) (bool, error) {

	// Skip deposit event, as outlined in the function documentation.
	if event.EventType == sidecartypes.DepositEventName {
		return true, nil
	}

	// For other event types, it is very important to validate the messages, otherwise invalid messages will not be seen
	// by the AnteHandler, which means we will not track the injected transactions correctly.

	parsedEvent, err := event.UnmarshalParsedEvent()
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal parsed event: %s; event: %s", err.Error(), event)
	}

	tx, err := h.txVerifier.TxDecode(rawTxBytes)
	if err != nil {
		return false, fmt.Errorf("failed to decode event transaction: %s", err.Error())
	}
	msgs := tx.GetMsgs()

	for _, msg := range msgs {
		m, ok := msg.(sdk.HasValidateBasic)
		if !ok {
			continue
		}

		if err := m.ValidateBasic(); err != nil {
			return false, nil // do not return the error, otherwise it takes priority over the boolean
		}
	}

	isAuthorizeEvent := event.EventType == sidecartypes.AuthorizeEventName
	err = h.authenticateTx(parsedEvent.Signer(), msgs, params, blockedBech32Addresses, isAuthorizeEvent)
	if err != nil {
		return false, nil // do not return the error, otherwise it takes priority over the boolean
	}

	return true, nil
}

// authenticateTx ensures that the msgs signer is the mapped Sequencer address of the sender.
//
// Special case: if messages are from an AuthorizeEvent we also check that they are listed in the allow list.
func (h *FuelSequencerProposalHandler) authenticateTx(
	sender string,
	msgs []sdk.Msg,
	params *bridgetypes.Params,
	blockedBech32Addresses map[string]bool,
	msgsAreFromAuthorizeEvent bool,
) error {

	// Generate the Sequencer address from the Ethereum address
	mappedSequencerAddr, err := h.bridgeKeeper.GenerateSequencerAddressFromEthereumAddress(sender)
	if err != nil {
		return bridgetypes.ErrCouldNotGenerateSequencerAddress.Wrapf("%v", err)
	}

	for _, msg := range msgs {

		// If message is from AuthorizeEvent, check that the message is allowed
		if msgsAreFromAuthorizeEvent && !params.IsAuthorizedMessage(msg) {
			return bridgetypes.ErrMsgNotAuthorizedOnSequencer.Wrapf("%s", sdk.MsgTypeURL(msg))
		}

		// Obtain the message signers using the proto signer annotations
		protoCodec, ok := h.cdc.(*codec.ProtoCodec)
		if !ok {
			return bridgetypes.ErrCodecIsNotSupported.Wrap(bridgetypes.ErrStrOnlyProtoCodecAllowed)
		}
		signers, _, err := protoCodec.GetMsgV1Signers(msg)
		if err != nil {
			return bridgetypes.ErrFailedToObtainMsgSigners.Wrapf("msg %s, err %v", sdk.MsgTypeURL(msg), err)
		}

		for _, signer := range signers {

			// Make sure that the message signer is equivalent to the mapped Sequencer address of the
			// sender on Ethereum. We also make sure that the signer is not part of a list of blocked addresses.
			signerAddress := sdk.AccAddress(signer).String()
			if mappedSequencerAddr.String() != signerAddress {
				return bridgetypes.ErrInvalidSigner.Wrapf(
					"expected %s, got %s", mappedSequencerAddr.String(), signerAddress,
				)
			}

			// Check if the bech32 signer is a blocked address.
			if blockedBech32Addresses[signerAddress] {
				return bridgetypes.ErrInvalidSigner.Wrapf(
					"signer %s is a blocked address ", signerAddress,
				)
			}
		}
	}

	return nil
}
