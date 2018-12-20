package main

import (
	"flag"
	"log"
	"net/http"
	"path"

	emotionengine "emotibot.com/emotigo/pkg/api/emotion-engine/v1"
)

func main() {
	flag.StringVar(&folder, "f", "./", "folder to scan for csv files. (default: ./)")
	flag.StringVar(&clientAddress, "addr", "localhost:8888", "emotion-engine address(default: localhost:8888")
	flag.Parse()
	var location http.FileSystem = http.Dir(folder)
	file, err := location.Open("/")
	if err != nil {
		log.Fatal("folder path "+folder+" error: ", err)
	}
	files, err := file.Readdir(0)
	if err != nil {
		log.Fatal("ReadDir error: ", err)
	}
	if len(files) == 0 {
		log.Fatal("dir is empty")
	}
	var model = emotionengine.Model{
		AppID:        "demo",
		IsAutoReload: true,
		Data:         make(map[string]emotionengine.Emotion, 0),
	}
	for _, f := range files {
		filename := f.Name()
		if path.Ext(filename) != ".csv" {
			continue
		}
		absFilePath := folder + "/" + filename
		emotion, err := emotionengine.CSVToEmotion(absFilePath, filename[:len(filename)-3])
		if err != nil {
			log.Println("csv file: ", f.Name(), "can not import: ", err)
		}
		model.Data[emotion.Name] = emotion

	}
	eClient := emotionengine.Client{
		Transport: http.DefaultClient,
		ServerURL: "http://" + clientAddress,
	}

	modelID, err := eClient.Train(model)
	if err != nil {
		log.Fatal("train failed, ", err)
	}

	log.Println("trained success, model_id: ", modelID, ", app_id: ", model.AppID)
	log.Println("trained emotions:")
	for emotionName, _ := range model.Data {
		log.Println(emotionName)
	}
}

var folder string
var clientAddress string
