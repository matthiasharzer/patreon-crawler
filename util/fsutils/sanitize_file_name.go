package fsutils

import (
	"regexp"
	"slices"
	"strings"
)

var fileNameWindowsReservedNames = []string{
	"CON", "PRN", "AUX", "NUL",
	"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
	"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
}

var fileNameInvalidChars = regexp.MustCompile(`[<>:"/\\|?*\x00]`)

func SanitizeFilename(name string) string {
	name = fileNameInvalidChars.ReplaceAllString(name, "_")
	name = strings.TrimRight(name, " .")
	upper := strings.ToUpper(name)
	if slices.Contains(fileNameWindowsReservedNames, upper) {
		name = "_" + name
	}
	if len(name) > 255 {
		name = name[:255]
	}
	if name == "" {
		name = "_"
	}
	return name
}
