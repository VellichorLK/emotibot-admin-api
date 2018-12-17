package main

import (
	"fmt"
	"net/http"
	"testing"

	"emotibot.com/emotigo/module/admin-api/util"
)

func TestHealthCheck(t *testing.T) {
	ret := healthCheck()
	if ret == nil {
		t.Error("Get nil health check")
	}
}

func TestSetRoute(t *testing.T) {
	modules = []*util.ModuleInfo{
		&util.ModuleInfo{
			ModuleName: "foo",
			EntryPoints: []util.EntryPoint{
				util.EntryPoint{
					AllowMethod: "GET",
					EntryPath:   "bar",
					Callback: func(w http.ResponseWriter, r *http.Request) {

					},
					Version: 2,
					Command: []string{},
				},
			},
		},
	}
	router := setRoute()
	fmt.Printf("%+v\n", router)
	route := router.Get("bar")
	if route == nil {
		t.Fatal("should be a route, but got nil")
	}
}
