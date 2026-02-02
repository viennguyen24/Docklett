package util

import (
	"fmt"
	"unicode/utf8"
)

func ReadSingleChar(source string, index int) (rune, error) {
	if index >= utf8.RuneCountInString(source) {
		return 0, fmt.Errorf("reading char at out of bounds index")
	}
	return []rune(source)[index], nil
}

func ReadSubstring(source string, start int, end int) (string, error) {
	if start < 0 || end > utf8.RuneCountInString(source) || start >= end {
		return "", fmt.Errorf("reading substring at out of bounds indices")
	}
	return string([]rune(source)[start:end]), nil
}