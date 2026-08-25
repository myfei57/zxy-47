package ns

import (
	"fmt"
	"strconv"
	"strings"
)

func ZoneID(shedID string, index int) string {
	return fmt.Sprintf("%s-Z%02d", shedID, index)
}

func ParseZoneID(zoneID string) (string, int, bool) {
	idx := strings.LastIndex(zoneID, "-Z")
	if idx <= 0 {
		return "", 0, false
	}
	number := zoneID[idx+2:]
	value, err := strconv.Atoi(number)
	if err != nil {
		return "", 0, false
	}
	return zoneID[:idx], value, true
}

func ZoneName(index int) string {
	return fmt.Sprintf("区%02d", index)
}

func BoundIndex(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
