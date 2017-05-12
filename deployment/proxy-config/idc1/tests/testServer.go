package main

import (
	//"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
)

type helloHandler struct {
	serverName string
}

func (h helloHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("req url: %s", r.URL)
	fmt.Println("Header:", r.Header)
	//	buf := new(bytes.Buffer)
	//	buf.ReadFrom(r.Body)
	//  fmt.Println("Body: ", buf.String())
	r.ParseMultipartForm(0)
	fmt.Println("USERID: ", r.FormValue("userid"))

	fmt.Println(r.FormValue("delete"))
	fmt.Fprintf(w, "This request goes to [%s]\n", h.serverName)
}

func main() {
	port := fmt.Sprintf(":%s", os.Args[1])
	serverName := fmt.Sprintf("%s", os.Args[2])
	fmt.Println("Listening on port", port, "/", serverName)

	err := http.ListenAndServe(port, helloHandler{serverName: serverName})
	log.Fatal(err)
}
