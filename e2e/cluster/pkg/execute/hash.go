package execute

import (
	"fmt"
	"os/exec"

	"golang.org/x/crypto/ssh"
)

// CalculateFileHash computes the SHA256 hash of a file, either locally or remotely.
// Returns the hash as a byte slice and any error encountered.
func CalculateFileHash(filepath string, client *ssh.Client) ([]byte, error) {
	cmd := "sha256sum " + filepath + " | cut -d' ' -f1"

	// Local file
	if client == nil {
		output, err := exec.Command("sh", "-c", cmd).Output()
		if err != nil {
			return nil, fmt.Errorf("failed to get file hash: %w", err)
		}
		return output, nil
	}

	// Remote file
	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()
	session.Stdout = nil // Disable stdout to capture output
	return session.Output(cmd)
}
