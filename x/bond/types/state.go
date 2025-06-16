package types

// NewState creates a new State instance.
func NewState(yieldMintHeight int64) State {
	return State{
		YieldMintHeight: yieldMintHeight,
	}
}

// DefaultState returns the default state.
func DefaultState() State {
	return NewState(0)
}
