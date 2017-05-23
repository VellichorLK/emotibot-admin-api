package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"emotibot.com/emotigo/experimental/scalable_service/web_service/handlers"
)

func main() {

	rabbitmqHost := os.Getenv("RABBITMQ_HOST")
	rabbitmqPort, err := strconv.Atoi(os.Getenv("RABBITMQ_PORT"))
	if err != nil {
		log.Fatalf("%s: %s", "1.Convert port error", err)
	}

	_, err = strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Fatalf("%s: %s", "2.Convert port error", err)
	}
	listenPort := os.Getenv("PORT")
	listenPort = ":" + listenPort

	handlers.LoadCfg("./html/serviceCfg.yaml")
	log.Printf("Loading cfg file completed.\n")

	handlers.RegisterAllHandlers()
	log.Printf("Register handlers completed.\n")

	handlers.InitController(rabbitmqHost, rabbitmqPort)
	http.Handle("/", http.FileServer(http.Dir("./html")))

	log.Printf("Start listening on %s port!", listenPort)
	http.ListenAndServe(listenPort, nil)

}
