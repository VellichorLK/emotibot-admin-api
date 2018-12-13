package main

import "testing"

func TestHealthCheck(t *testing.T) {
	ret := healthCheck()
	if ret == nil {
		t.Error("Get nil health check")
	}
}
