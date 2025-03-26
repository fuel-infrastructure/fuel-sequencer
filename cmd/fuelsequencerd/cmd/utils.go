package cmd

import (
	"fmt"
	"strconv"
)

func validatePort(port string) error {
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("invalid port format: %w", err)
	}
	if portNum < 1 || portNum > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", portNum)
	}
	return nil
}
