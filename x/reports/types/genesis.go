package types

// this line is used by starport scaffolding # genesis/types/import

// DefaultIndex is the default global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:          DefaultParams(),
		SlashReportList: []SlashReport{},
		// this line is used by starport scaffolding # genesis/types/default
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {

	// Check for duplicated heights in SlashReportList
	uniqueSlashReports := make(map[uint64]bool)

	for _, sr := range gs.SlashReportList {

		// verify slash report is unique
		if _, ok := uniqueSlashReports[sr.Height]; ok {
			return ErrSlashReportNotUnique.Wrapf("slash report not unique at height %d", sr.Height)
		}

		uniqueSlashReports[sr.Height] = true
	}
	// this line is used by starport scaffolding # genesis/types/validate

	return gs.Params.Validate()
}
