package utils

import (
	"fmt"
	"math"
	"strings"
	"time"
)

func Max(a int, b int) int {
	if a > b {
		return a
	}

	return b
}

func Round(val float64, digits int) float64 {
	multiplier := math.Pow10(digits)

	val = math.Round(val * multiplier)

	return val / multiplier
}

/* https://yourbasic.org/golang/formatting-byte-size-to-human-readable-format/ */
func FormatBytes(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}

	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}

func FormatDurationPretty(d time.Duration) string {
	strParts := []string{}

	if d > 24*time.Hour {
		days := d / (24 * time.Hour)
		d = d % (24 * time.Hour)

		strParts = append(strParts, fmt.Sprintf("%dd", days))
	}

	if d > 1*time.Hour {
		hours := d / time.Hour
		d = d % time.Hour

		strParts = append(strParts, fmt.Sprintf("%dh", hours))
	}

	if d > 1*time.Minute {
		minutes := d / time.Minute
		d = d % time.Minute

		strParts = append(strParts, fmt.Sprintf("%dm", minutes))
	}

	if d > 1*time.Second {
		seconds := d / time.Second

		strParts = append(strParts, fmt.Sprintf("%ds", seconds))
	}

	if len(strParts) == 0 {
		strParts = append(strParts, "0s")
	}

	return strings.Join(strParts, " ")
}
