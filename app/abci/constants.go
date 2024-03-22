package abci

var (
	// MsgSupplyDeltaGasLimit is the default gas limit set to MsgSupplyDelta when injected. This value is set to zero
	// because inside the message handler we are overriding with an infinite gas meter to make sure that MsgSupplyDelta
	// does not fail due to insufficient gas.
	MsgSupplyDeltaGasLimit = uint64(0)
)
