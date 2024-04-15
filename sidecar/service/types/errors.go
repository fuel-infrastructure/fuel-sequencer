package types

import "strings"

const ErrBlockDoesNotExist = "block does not exist yet"
const ErrSidecarFallenBehindWithAcceptableDelay = "sidecar has fallen behind the network with acceptable delay"

// IsErrorFatal is a helper which determines whether the error raised by the Sidecar should halt the block production
// on the Sequencer. If true is returned, the Sequencer should halt block production, otherwise, the Sequencer can
// proceed with block production.
func IsErrorFatal(err error) bool {
	// This is a list of Sidecar errors that are not fatal.
	acceptableErrors := []string{

		// Sidecar is out-of-sync with Ethereum for a few amount of blocks.
		ErrSidecarFallenBehindWithAcceptableDelay,

		// Queried block does not exist yet on Ethereum.
		ErrBlockDoesNotExist,
	}

	// If we match one of the acceptable errors the error is not fatal.
	for _, errMsg := range acceptableErrors {
		if strings.Contains(err.Error(), errMsg) {
			return false
		}
	}

	// If an error is not acceptable, then it must be fatal.
	return true
}
