package data

import (
	"fmt"
	"testing"
)

func TestStopWords(t *testing.T) {
	fmt.Println(InitializeWord2Vec("../../"))
	Word2Vec.Range(func(key, value interface{}) bool {
		if key == "老翁" {
			fmt.Println(key, value)
		}
		return true
	})
}
