package testsuite

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	cmcfg "github.com/cometbft/cometbft/config"
	cmos "github.com/cometbft/cometbft/libs/os"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/privval"
	dbm "github.com/cosmos/cosmos-db"
	sdkcrypto "github.com/cosmos/cosmos-sdk/crypto"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/server"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	txsigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	fuelsequencerapp "github.com/fuel-infrastructure/fuel-sequencer/app"
	sidecarconfig "github.com/fuel-infrastructure/fuel-sequencer/sidecar/config"
)

type validator struct {
	chain        *Chain
	index        int
	Moniker      string
	mnemonic     string
	keyRecord    keyring.Record
	privateKey   cryptotypes.PrivKey
	consensusKey privval.FilePVKey
	NodeKey      p2p.NodeKey

	// FuelSequencer ports set during startup.
	hostRPCPort     string
	hostAPIPort     string
	hostGRPCPort    string
	sidecarGRPCPort string
}

func (v *validator) InstanceName() string {
	return fmt.Sprintf("%s%d", v.Moniker, v.index)
}

func (v *validator) ConfigDir() string {
	return filepath.Join(v.chain.ConfigDir(), v.InstanceName())
}

func (v *validator) createConfig() error {
	p := path.Join(v.ConfigDir(), "config")
	return os.MkdirAll(p, 0755)
}

func (v *validator) init() error {
	if err := v.createConfig(); err != nil {
		return err
	}

	serverCtx := server.NewDefaultContext()
	config := serverCtx.Config

	config.SetRoot(v.ConfigDir())
	config.Moniker = v.Moniker

	genDoc, err := getGenDoc(v.ConfigDir())
	if err != nil {
		return err
	}

	db := dbm.NewMemDB()
	appOpts := simtestutil.AppOptionsMap{sidecarconfig.FlagSidecarEnabled: false}
	app, err := fuelsequencerapp.NewFuelSequencerApp(log.NewNopLogger(), db, nil, true, false, appOpts)
	if err != nil {
		panic(err)
	}
	genesisState := app.DefaultGenesis()
	if err != nil {
		panic(err)
	}
	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	if err != nil {
		panic(err)
	}

	appGenesis := &genutiltypes.AppGenesis{
		AppName:       fuelsequencerapp.Name,
		AppVersion:    "v1", // TODO: fuelsequencerapp.Version,
		GenesisTime:   genDoc.GenesisTime,
		ChainID:       v.chain.id,
		InitialHeight: genDoc.InitialHeight,
		AppHash:       genDoc.AppHash,
		AppState:      stateBytes,
		Consensus: &genutiltypes.ConsensusGenesis{
			Validators: nil,
		},
	}

	if err = genutil.ExportGenesisFile(appGenesis, config.GenesisFile()); err != nil {
		return fmt.Errorf("failed to export app genesis state: %w", err)
	}

	cmcfg.WriteConfigFile(filepath.Join(config.RootDir, "config", "config.toml"), config)
	return nil
}

func (v *validator) createNodeKey() error {
	serverCtx := server.NewDefaultContext()
	config := serverCtx.Config

	config.SetRoot(v.ConfigDir())
	config.Moniker = v.Moniker

	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	if err != nil {
		return err
	}

	v.NodeKey = *nodeKey
	return nil
}

func (v *validator) createConsensusKey() error {
	serverCtx := server.NewDefaultContext()
	config := serverCtx.Config

	config.SetRoot(v.ConfigDir())
	config.Moniker = v.Moniker

	pvKeyFile := config.PrivValidatorKeyFile()
	if err := cmos.EnsureDir(filepath.Dir(pvKeyFile), 0777); err != nil {
		return err
	}

	pvStateFile := config.PrivValidatorStateFile()
	if err := cmos.EnsureDir(filepath.Dir(pvStateFile), 0777); err != nil {
		return err
	}

	filePV := privval.LoadOrGenFilePV(pvKeyFile, pvStateFile)
	v.consensusKey = filePV.Key

	return nil
}

func (v *validator) createKeyFromMnemonic(name, mnemonic string, passphrase string) error {
	kb, err := keyring.New(keyringAppName, keyring.BackendTest, v.ConfigDir(), nil, Cdc)
	if err != nil {
		return err
	}

	keyringAlgos, _ := kb.SupportedAlgorithms()
	algo, err := keyring.NewSigningAlgoFromString(string(hd.Secp256k1Type), keyringAlgos)
	if err != nil {
		return err
	}

	record, err := kb.NewAccount(name, mnemonic, passphrase, sdk.FullFundraiserPath, algo)
	if err != nil {
		return err
	}

	privKeyArmor, err := kb.ExportPrivKeyArmor(name, keyringPassphrase)
	if err != nil {
		return err
	}

	privKey, _, err := sdkcrypto.UnarmorDecryptPrivKey(privKeyArmor, keyringPassphrase)
	if err != nil {
		return err
	}

	v.keyRecord = *record
	v.mnemonic = mnemonic
	v.privateKey = privKey

	return nil
}

func (v *validator) createKey(name string) error { //nolint:unused
	mnemonic, err := createMnemonic()
	if err != nil {
		return err
	}

	return v.createKeyFromMnemonic(name, mnemonic, "")
}

func (v *validator) BuildCreateValidatorMsg(moniker string, amount sdk.Coin) (sdk.Msg, error) {
	description := stakingtypes.NewDescription(moniker, "", "", "", "")
	commissionRates := stakingtypes.CommissionRates{
		Rate:          math.LegacyMustNewDecFromStr("0.1"),
		MaxRate:       math.LegacyMustNewDecFromStr("0.2"),
		MaxChangeRate: math.LegacyMustNewDecFromStr("0.01"),
	}

	// get the initial validator min self delegation
	minSelfDelegation, _ := math.NewIntFromString("1")

	valPubKey, err := cryptocodec.FromTmPubKeyInterface(v.consensusKey.PubKey)
	if err != nil {
		return nil, err
	}

	addr, err := v.keyRecord.GetAddress()
	if err != nil {
		return nil, err
	}

	return stakingtypes.NewMsgCreateValidator(
		sdk.ValAddress(addr).String(),
		valPubKey,
		amount,
		description,
		commissionRates,
		minSelfDelegation,
	)
}

func (v *validator) SignMsg(msgs ...sdk.Msg) (*sdktx.Tx, error) {
	txBuilder := encodingConfig.TxConfig.NewTxBuilder()

	if err := txBuilder.SetMsgs(msgs...); err != nil {
		return nil, err
	}

	txBuilder.SetMemo(fmt.Sprintf("%s@%s:26656", v.NodeKey.ID(), v.InstanceName()))
	txBuilder.SetFeeAmount(sdk.NewCoins())
	txBuilder.SetGasLimit(200000)

	signerData := authsigning.SignerData{
		Address:       v.pubKey().Address().String(),
		ChainID:       v.chain.id,
		AccountNumber: 0,
		Sequence:      0,
		PubKey:        v.pubKey(),
	}

	// For SIGN_MODE_DIRECT, calling SetSignatures calls setSignerInfos on
	// TxBuilder under the hood, and SignerInfos is needed to generate the sign
	// bytes. This is the reason for setting SetSignatures here, with a nil
	// signature.
	//
	// Note: This line is not needed for SIGN_MODE_LEGACY_AMINO, but putting it
	// also doesn't affect its generated sign bytes, so for code's simplicity
	// sake, we put it here.
	sig := txsigning.SignatureV2{
		PubKey: v.pubKey(),
		Data: &txsigning.SingleSignatureData{
			SignMode:  txsigning.SignMode_SIGN_MODE_DIRECT,
			Signature: nil,
		},
		Sequence: 0,
	}

	if err := txBuilder.SetSignatures(sig); err != nil {
		return nil, err
	}

	bytesToSign, err := authsigning.GetSignBytesAdapter(
		context.Background(), // TODO: is this fine?
		encodingConfig.TxConfig.SignModeHandler(),
		txsigning.SignMode_SIGN_MODE_DIRECT,
		signerData,
		txBuilder.GetTx(),
	)
	if err != nil {
		return nil, err
	}

	sigBytes, err := v.privateKey.Sign(bytesToSign)
	if err != nil {
		return nil, err
	}

	sig = txsigning.SignatureV2{
		PubKey: v.pubKey(),
		Data: &txsigning.SingleSignatureData{
			SignMode:  txsigning.SignMode_SIGN_MODE_DIRECT,
			Signature: sigBytes,
		},
		Sequence: 0,
	}
	if err := txBuilder.SetSignatures(sig); err != nil {
		return nil, err
	}

	signedTx := txBuilder.GetTx()
	bz, err := encodingConfig.TxConfig.TxEncoder()(signedTx)
	if err != nil {
		return nil, err
	}

	return decodeTx(bz)
}

func (v *validator) keyring() (keyring.Keyring, error) {
	return keyring.New(keyringAppName, keyring.BackendTest, v.ConfigDir(), nil, Cdc)
}

func (v *validator) Address() sdk.AccAddress {
	addr, err := v.keyRecord.GetAddress()
	if err != nil {
		panic(err)
	}

	return addr
}

func (v *validator) validatorAddress() sdk.ValAddress {
	return sdk.ValAddress(v.Address())
}

func (v *validator) pubKey() cryptotypes.PubKey {
	pubKey, err := v.keyRecord.GetPubKey()
	if err != nil {
		panic(err)
	}

	return pubKey
}
