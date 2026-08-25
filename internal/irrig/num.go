package irrig

import "strconv"

func formatNumber(value float64) string {
	return strconv.FormatFloat(value, 'f', 2, 64)
}
