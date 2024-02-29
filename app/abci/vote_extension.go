package abci

/**
TODO: Types inside this file should be replaced by the types required by our application
*/

// CustomOracleVoteExtension defines the canonical vote extension structure.
type CustomOracleVoteExtension struct {
	Height int64 // this is technically optional
	Data   CustomData
}

// CustomData this is just to show that the vote extension above does not need to be in bytes
type CustomData struct {
	Data []byte
	Msgs string
}
