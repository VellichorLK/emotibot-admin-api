package fileservice

import (
	"bytes"
	"testing"
)

func TestPut(T *testing.T) {
	err := Init(
		"172.16.101.98:9090",
		"24WTCKX23M0MRI6BA72Y",
		"6mxoQMf1Lha2q1H3fp3fjNHEZRJJ7LvbC5k3+CQX",
		false)
	if err != nil {
		T.Fatalf("Cannot init minio, %s", err.Error())
	}

	content := []byte("temporary file's content")
	reader := bytes.NewReader(content)
	err = AddFile("testing", "testFile", reader)
	if err != nil {
		T.Fatalf("Cannot add file, %s", err.Error())
	}
}
