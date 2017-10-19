package main

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func BenchmarkUnmarshal(b *testing.B) {
	b1 := new(bytes.Buffer)
	f, _ := os.Open("test.json")
	io.Copy(b1, f)
	f.Close()

	task := string(b1.Bytes())

	for i := 0; i < b.N; i++ {
		parseJSON(&task)
	}

}
