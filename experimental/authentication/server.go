package main

import (
	"emotibot.com/emotigo/experimental/authentication/auth"
	"fmt"
	"log"
	"net/http"
)

func main() {
	c := auth.GetConfig()
	log.Printf("config: %s", c)
	err := auth.SetRoute(c)
	if err != nil {
		log.Fatalf("set route failed. %s", err.Error())
	}
	p := fmt.Sprintf(":%s", c.ListenPort)
	http.ListenAndServe(p, nil)
	log.Println("bye!")
}
