package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func main() {
	rand.Seed(time.Now().Unix())
	m := http.NewServeMux()
	m.HandleFunc("/vip/irobot/get-questions.action", fakeHandler)
	server := http.Server{
		Addr:    ":15801",
		Handler: m,
	}
	log.Println("server starting...")
	log.Println(server.ListenAndServe())
}

func fakeHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("receive request from %s\n", r.RemoteAddr)
	if (rand.Intn(99) % 100) != 0 {
		time.Sleep(time.Duration(100) * time.Millisecond)
	} else {
		time.Sleep(time.Duration(4000) * time.Millisecond)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Success")
	log.Println("request handled")
	return
}
