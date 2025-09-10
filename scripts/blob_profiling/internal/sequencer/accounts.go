package sequencer

type Account struct {
	Name     string
	Address  string
	Mnemonic string

	Sequence uint64
}

var Accounts = map[string]Account{
	"alice": {
		Name:     "alice",
		Address:  "fuelsequencer1vtfzrk6f4m6kxt6ehyqt9j5su5hvcz5q3dmlsm",
		Mnemonic: "dinner crash nurse casino baby fold race cheese elite column sausage sleep close royal rain over mechanic minimum outdoor conduct cash wagon frog evidence",
	},
	"bob": {
		Name:     "bob",
		Address:  "fuelsequencer163rsv65t4893t2rz5rmda9sly7lgdlq2jgr36m",
		Mnemonic: "gaze drama excess raven follow antenna swallow beef upper myself question pitch course ill adult century crisp ice rough match praise sing unveil vintage",
	},
	"carol": {
		Name:     "carol",
		Address:  "fuelsequencer1n79wsstpakv0gw2efmruf9x8xs9m4rqfazfu8g",
		Mnemonic: "bar describe panda mosquito quiz room daring round nurse disagree swallow frown hat repeat recall flight skin sketch volume dutch range grunt assist nerve",
	},
	"dexter": {
		Name:     "dexter",
		Address:  "fuelsequencer1r8aaf8fjcft7h7tupnafyv0kzk0342gnjls2pu",
		Mnemonic: "bonus clinic owner choose grief soda ride divorce album oval tone mixed mechanic coin defense wonder tumble vault sorry great hover neither security amazing",
	},
	"eve": {
		Name:     "eve",
		Address:  "fuelsequencer17w0adeg64ky0daxwd2ugyuneellmjgnx5dpmtz",
		Mnemonic: "dinner crash nurse casino baby fold race cheese elite column sausage sleep close royal rain over mechanic minimum outdoor conduct cash wagon frog evidence",
	},
}
