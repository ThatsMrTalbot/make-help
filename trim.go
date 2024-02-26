package main

import "strings"

// trimLeadingSpaces mutates the provided slice.
//
// It counts the leading spaces on the first line, then trims that number
// of spaces from all lines
func trimLeadingSpaces(lines []string) {
	if len(lines) == 0 {
		return
	}

	count := countLeadingSpaces(lines[0])
	prefix := strings.Repeat(" ", count)

	for i, c := range lines {
		lines[i] = strings.TrimPrefix(c, prefix)
	}
}

func countLeadingSpaces(line string) int {
	for i, c := range line {
		if c != ' ' {
			return i
		}
	}
	return 0
}
