package execute

import "fmt"

// WithSudo wraps a command to be executed with sudo privileges using the provided password
func WithSudo(cmd string, password string) string {
	return fmt.Sprintf("echo '%s' | sudo -S %s", password, cmd)
}
