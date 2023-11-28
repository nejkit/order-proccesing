package util

import "strings"

func Min(a, b float64) float64 {
	if a <= b {
		return a
	}
	return b
}

func ParseCurrencyFromDirection(pair string, dir int) string {
	return strings.Split(pair, "/")[dir-1]
}
