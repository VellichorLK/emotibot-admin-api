package qi

import (
	"os"
	"testing"

	"emotibot.com/emotigo/pkg/logger"
)

func TestMain(m *testing.M) {
	// Avoid unnecessary printing in testing
	logger.SetLevel("ERROR")
	os.Exit(m.Run())
}

func BackupPointers(ptrs ...interface{}) func() {
	var tmp = make([]interface{}, len(ptrs))
	for i, ptr := range ptrs {
		tmp[i] = ptr
	}
	return func() {
		for i, t := range tmp {
			ptrs[i] = t
		}
	}
}
