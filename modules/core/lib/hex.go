package lib

import (
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidHexChar = errors.New("invalid hex character")

var encodeMap = map[rune]rune{
	'0': 'g',
	'1': 'h',
}

var decodeMap = map[rune]rune{
	'g': '0',
	'h': '1',
}

// HexEncode converts a number to a custom hex string.
func HexEncode(n int64) string {
	hex := fmt.Sprintf("%x", n)

	var result strings.Builder
	for _, c := range hex {
		if mapped, ok := encodeMap[c]; ok {
			result.WriteRune(mapped)
		} else {
			result.WriteRune(c)
		}
	}

	return result.String()
}

// HexDecode converts a custom hex string back to a number.
func HexDecode(s string) (int64, error) {
	if s == "" {
		return 0, ErrInvalidHexChar
	}

	var result strings.Builder
	for _, c := range s {
		if mapped, ok := decodeMap[c]; ok {
			result.WriteRune(mapped)
		} else {
			result.WriteRune(c)
		}
	}

	var n int64
	_, err := fmt.Sscanf(result.String(), "%x", &n)
	if err != nil {
		return 0, ErrInvalidHexChar
	}

	return n, nil
}
