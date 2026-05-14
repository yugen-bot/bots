package utils

import "fmt"

func FormatMinutes(minutes int) string {
	hours := minutes / 60
	mins := minutes % 60
	switch {
	case hours > 0 && mins > 0:
		return fmt.Sprintf(
			"%d hour%s %d minute%s",
			hours, PluralS(hours),
			mins, PluralS(mins),
		)
	case hours > 0:
		return fmt.Sprintf("%d hour%s", hours, PluralS(hours))
	default:
		return fmt.Sprintf("%d minute%s", mins, PluralS(mins))
	}
}

func PluralS(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
