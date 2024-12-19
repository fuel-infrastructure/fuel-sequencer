package cluster

// import (
// 	"fmt"
// 	"strings"
// )

// func cleanDestinations(connections []connection) error {
// 	for _, conn := range connections {
// 		cmd := fmt.Sprintf("pkill -f %s", binaryName)
// 		if err := executeRemote(conn.SSH, cmd); err != nil {
// 			// Ignore "no process found" errors
// 			if !strings.Contains(err.Error(), "no process found") {
// 				return fmt.Errorf("failed to clean destination %s: %w", conn.destination.host, err)
// 			}
// 		}
// 	}
// 	return nil
// }
