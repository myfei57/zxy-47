package console

import "time"

func unixNano() int64 {
	return time.Now().UnixNano()
}
