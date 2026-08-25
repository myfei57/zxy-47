package fert

import "strconv"

func ftoa(value float64) string {
	return strconv.FormatFloat(value, 'f', 2, 64)
}

func itoa(value int64) string {
	return strconv.FormatInt(value, 10)
}
