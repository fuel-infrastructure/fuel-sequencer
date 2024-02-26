package config

import (
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/joho/godotenv"
)

const (
	ETH_NODE_API_KEY = "ETH_NODE_API_KEY"

	CONTRACT_ADDRESS_KEY = "CONTRACT_ADDRESS_KEY"

	ETH_START_QUERY_BLOCK_KEY = "ETH_START_QUERY_BLOCK_KEY"
)

// GetEnvironmentalVariables returns the env variables in the .env file
func GetEnvironmentalVariables() (string, string, *big.Int) {

	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	ethNodeAPI := os.Getenv(ETH_NODE_API_KEY)
	if ethNodeAPI == "" {
		panic(fmt.Sprintf("environment variable %s was not set", ETH_NODE_API_KEY))
	}

	contractAddress := os.Getenv(CONTRACT_ADDRESS_KEY)
	if contractAddress == "" {
		panic(fmt.Sprintf("environment variable %s was not set", CONTRACT_ADDRESS_KEY))
	}

	ethStartBlock := new(big.Int)
	ethStartBlockStr := os.Getenv(ETH_START_QUERY_BLOCK_KEY)
	if ethStartBlockStr == "" {
		ethStartBlock = big.NewInt(0)
	} else {
		_, ok := ethStartBlock.SetString(ethStartBlockStr, 10) // Base 10 for decimal values
		if !ok {
			log.Printf("Warning: Invalid %s value, using default value. Got: %s", ETH_START_QUERY_BLOCK_KEY, ethStartBlockStr)
		}
	}

	return ethNodeAPI, contractAddress, ethStartBlock
}
