package execute

import "strings"

// isProgressLine determines if a command output line represents a progress update.
// This function identifies lines that contain progress information typically generated
// by tools like rsync, wget, curl, and other transfer/download utilities.
//
// Progress lines are characterized by:
//   - Percentage indicators (%)
//   - Transfer speed indicators (MB/s, KB/s, etc.)
//   - Transfer completion indicators (xfer#, to-check, etc.)
//
// Examples of progress lines:
//   - "5046272   5%    4.75MB/s   00:00:19"
//   - "100204544 100%    1.64MB/s   00:00:00 (xfer#1, to-check=0/1)"
//   - "Downloading: 45% [2.3MB/s] [00:15<00:12]"
//
// This detection allows for structured logging of progress updates while
// maintaining regular logging for other command output.
func isProgressLine(line string) bool {
	// Must contain a percentage indicator
	if !strings.Contains(line, "%") {
		return false
	}

	// Must contain at least one of the common progress indicators
	progressIndicators := []string{
		"MB/s", "KB/s", "GB/s", // Transfer speeds
		"xfer#", "to-check", // rsync completion indicators
		"ETA", "ETA:", // Estimated time remaining
		"Downloading:", // Download progress
		"Uploading:",   // Upload progress
		"Progress:",    // Generic progress indicator
	}

	for _, indicator := range progressIndicators {
		if strings.Contains(line, indicator) {
			return true
		}
	}

	return false
}
