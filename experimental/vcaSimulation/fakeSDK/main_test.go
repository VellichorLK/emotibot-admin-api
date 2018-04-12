package main

import (
	"testing"

	"net/http"
	"net/http/httptest"
)

func TestFakeHandler(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(fakeHandler))
	defer server.Close()
	// resp, err := http.Get(server.URL)
	// if err != nil {
	// 	t.Fatal(err)
	// }

}
