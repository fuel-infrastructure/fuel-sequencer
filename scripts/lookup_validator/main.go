package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/btcsuite/btcutil/bech32"
)

// Validator represents a validator from the API response
type Validator struct {
	OperatorAddress string `json:"operator_address"`
	ConsensusPubkey struct {
		Type string `json:"@type"`
		Key  string `json:"key"`
	} `json:"consensus_pubkey"`
	Description struct {
		Moniker string `json:"moniker"`
	} `json:"description"`
	Status string `json:"status"`
	Tokens string `json:"tokens"`
	Jailed bool   `json:"jailed"`
}

// ValidatorsResponse represents the API response
type ValidatorsResponse struct {
	Validators []Validator `json:"validators"`
}

// convertPubkeyToValcons converts a base64 consensus public key to valcons address
func convertPubkeyToValcons(pubkeyB64, prefix string) (string, error) {
	// Decode base64 public key
	pubkeyBytes, err := base64.StdEncoding.DecodeString(pubkeyB64)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 pubkey: %w", err)
	}

	// Take SHA-256 hash of the public key
	hash := sha256.Sum256(pubkeyBytes)

	// Take first 20 bytes
	address := hash[:20]

	// Convert to 5-bit groups for bech32
	conv, err := bech32.ConvertBits(address, 8, 5, true)
	if err != nil {
		return "", fmt.Errorf("failed to convert bits: %w", err)
	}

	// Encode with bech32
	valcons, err := bech32.Encode(prefix, conv)
	if err != nil {
		return "", fmt.Errorf("failed to encode bech32: %w", err)
	}

	return valcons, nil
}

// convertOperatorToAccount converts operator address to account address
func convertOperatorToAccount(operatorAddr string) string {
	return strings.Replace(operatorAddr, "fuelsequencervaloper", "fuelsequencer", 1)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <valcons_address>")
		fmt.Println("Example: go run main.go fuelsequencervalcons109s7d24matadmlpckg39qkjyy8j79kt02pw7dv")
		os.Exit(1)
	}

	targetValcons := os.Args[1]
	restURL := "https://rest-fuel-seq.simplystaking.xyz/cosmos/staking/v1beta1/validators"

	fmt.Printf("🔍 Looking for validator with valcons: %s\n\n", targetValcons)

	// Fetch validators from REST API
	resp, err := http.Get(restURL)
	if err != nil {
		fmt.Printf("❌ Error fetching validators: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ Error reading response: %v\n", err)
		os.Exit(1)
	}

	var validatorsResp ValidatorsResponse
	err = json.Unmarshal(body, &validatorsResp)
	if err != nil {
		fmt.Printf("❌ Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📊 Total validators found: %d\n", len(validatorsResp.Validators))
	fmt.Println("🔄 Converting consensus public keys to valcons addresses...\n")

	found := false
	for i, validator := range validatorsResp.Validators {
		if validator.ConsensusPubkey.Key == "" {
			fmt.Printf("⚠️  Validator %d: No consensus pubkey found\n", i+1)
			continue
		}

		// Convert consensus pubkey to valcons address
		valcons, err := convertPubkeyToValcons(validator.ConsensusPubkey.Key, "fuelsequencervalcons")
		if err != nil {
			fmt.Printf("⚠️  Validator %d: Error converting pubkey: %v\n", i+1, err)
			continue
		}

		fmt.Printf("🔗 %s -> %s\n", validator.Description.Moniker, valcons)

		// Check if this matches our target
		if valcons == targetValcons {
			found = true
			fmt.Printf("\n🎯 ✅ MATCH FOUND!\n")
			fmt.Printf("═══════════════════════════════════════\n")
			fmt.Printf("📛 Moniker:        %s\n", validator.Description.Moniker)
			fmt.Printf("🏢 Operator Addr:  %s\n", validator.OperatorAddress)
			fmt.Printf("👤 Account Addr:   %s\n", convertOperatorToAccount(validator.OperatorAddress))
			fmt.Printf("🔐 Valcons Addr:   %s\n", valcons)
			fmt.Printf("📊 Status:         %s\n", validator.Status)
			fmt.Printf("💰 Tokens:         %s\n", validator.Tokens)
			fmt.Printf("⛓️  Jailed:         %t\n", validator.Jailed)
			fmt.Printf("🔑 Consensus Key:  %s...\n", validator.ConsensusPubkey.Key[:20])
			fmt.Printf("═══════════════════════════════════════\n")
			break
		}
	}

	if !found {
		fmt.Printf("\n❌ No validator found with valcons address: %s\n", targetValcons)
		fmt.Println("\n💡 Double-check the valcons address or try with a different one.")
	}
}
