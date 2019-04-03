package qi

import "time"

var unix = func() int64 {
	return time.Now().Unix()
}
