package util

import (
	"testing"
)

func TestRandomString(t *testing.T) {
	s := GenRandomString(10)
	if len(s) != 10 {
		t.Errorf("Except len 10, get %d", len(s))
	}
}
