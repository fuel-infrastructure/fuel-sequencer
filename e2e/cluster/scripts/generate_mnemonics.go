package main

import (
	"flag"
	"fmt"

	"github.com/cosmos/go-bip39"
)

func main() {
	count := flag.Int("count", 60, "number of mnemonics to generate")
	bits := flag.Int("bits", 128, "entropy bits (128=12 words, 256=24 words)")
	flag.Parse()

	fmt.Println("// Generated valid BIP-39 mnemonics")
	for i := 0; i < *count; i++ {
		entropy, err := bip39.NewEntropy(*bits)
		if err != nil {
			panic(fmt.Sprintf("failed to generate entropy: %v", err))
		}

		mnemonic, err := bip39.NewMnemonic(entropy)
		if err != nil {
			panic(fmt.Sprintf("failed to generate mnemonic: %v", err))
		}

		// Verify the mnemonic is valid
		if !bip39.IsMnemonicValid(mnemonic) {
			panic(fmt.Sprintf("generated invalid mnemonic: %s", mnemonic))
		}

		fmt.Printf("\t\t%q,\n", mnemonic)
	}
}
