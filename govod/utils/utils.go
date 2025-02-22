package utils

import (
	"strconv"
	"strings"
)

func GetBandwidth(bitrate string) int {
	valueStr := strings.TrimRight(bitrate, "kM")
	unit := strings.TrimLeft(bitrate, valueStr)
	value, _ := strconv.Atoi(valueStr)

	switch unit {
	case "k":
		return value * 1000
	case "M":
		return value * 1000000
	default:
		return value
	}
}
