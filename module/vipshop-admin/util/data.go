package util

import (
	"fmt"
	"time"
)

//JSONUnixTime are use for formatting to Unix Time Mill Second
type JSONUnixTime time.Time

// MarshalJSON is for JSON Marshal usage
func (t JSONUnixTime) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("%d", time.Time(t).UnixNano()/1000000)
	return []byte(stamp), nil
}
