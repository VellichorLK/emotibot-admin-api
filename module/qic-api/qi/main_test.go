package qi

import (
	"os"
	"reflect"
	"testing"

	"emotibot.com/emotigo/pkg/logger"
)

func TestMain(m *testing.M) {
	// Avoid unnecessary printing in testing
	logger.SetLevel("ERROR")
	os.Exit(m.Run())
}

func BackupPointers(funcPtrs ...interface{}) func() {
	var oldValues = make([]reflect.Value, len(funcPtrs))
	for i, ptr := range funcPtrs {
		oldFunc := reflect.ValueOf(ptr).Elem().Interface()
		oldValue := reflect.ValueOf(oldFunc)
		oldValues[i] = oldValue
	}
	return func() {
		for i, value := range oldValues {
			reflect.ValueOf(funcPtrs[i]).Elem().Set(value)
		}
	}
}
