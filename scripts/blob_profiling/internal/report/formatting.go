package report

import (
	"fmt"
	"time"

	"github.com/fuel-infrastructure/blob-storage/pkg/size"
)

// formatDuration returns a human-readable duration string
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.0fms", float64(d.Nanoseconds())/1e6)
	} else if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	} else {
		return fmt.Sprintf("%.1fh", d.Hours())
	}
}

// formatBytesPerSecond returns a human-readable bytes per second string
func formatBytesPerSecond(bytesPerSec float64) string {
	if bytesPerSec < size.KiB {
		return fmt.Sprintf("%.0f B/s", bytesPerSec)
	} else if bytesPerSec < size.MiB {
		return fmt.Sprintf("%.1f KiB/s", float64(bytesPerSec)/1024)
	} else if bytesPerSec < size.GiB {
		return fmt.Sprintf("%.1f MiB/s", float64(bytesPerSec)/(1024*1024))
	} else {
		return fmt.Sprintf("%.1f GiB/s", float64(bytesPerSec)/(1024*1024*1024))
	}
}
