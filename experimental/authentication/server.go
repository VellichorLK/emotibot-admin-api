package main

import (
	"emotibot.com/emotigo/experimental/authentication/auth"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func main() {
	auth.LogInit(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
	c := auth.GetConfig()
	auth.LogInfo.Printf("config: %s", c)
	err := auth.SetRoute(c)
	if err != nil {
		auth.LogError.Fatalf("set route failed. %s", err.Error())
	}
	p := fmt.Sprintf(":%s", c.ListenPort)
	http.ListenAndServe(p, nil)
	auth.LogInfo.Println("bye!")
}
