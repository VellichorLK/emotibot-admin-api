package imagesManager

import (
	"testing"
)

func TestUniqueString(t *testing.T) {
	length := 10
	name1 := GetUniqueString(length)
	name2 := GetUniqueString(length)

	if name1 == name2 {
		t.Fatal("name1 has the same name as name 2 " + name1)
	}
	if len(name1) != length {
		t.Fatalf("name1 has len %d, not %d\n", len(name1), length)
	}

	if len(name2) != length {
		t.Fatalf("name2 has len %d, not %d\n", len(name2), length)
	}

}
