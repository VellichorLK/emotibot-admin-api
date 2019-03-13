package general

import (
	"testing"
)

func TestStringsToRunes(t *testing.T) {
	ss := []string{
		"aa",
		"bb",
	}
	words := StringsToRunes(ss)

	if len(words) != len(ss) {
		t.Error("tranforms strings to runes failed")
	}
}
