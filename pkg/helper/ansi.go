// Package helper provides shared utility functions.
package helper

import "strings"

// StripANSI removes ANSI escape sequences from a string.
func StripANSI(s string) string {
	var result strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' {
			j := i + 1
			if j < len(s) && s[j] == '[' {
				j++
				for j < len(s) && (s[j] < 'a' || s[j] > 'z') && (s[j] < 'A' || s[j] > 'Z') {
					j++
				}
				j++
			}
			i = j
		} else {
			result.WriteByte(s[i])
			i++
		}
	}
	return result.String()
}
