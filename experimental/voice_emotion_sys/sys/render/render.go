package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"

	"handlers"
)

var envs = make(map[string]string)
var variableLists = [...]string{"RABBITMQ_HOST", "RABBITMQ_PORT", "DB_HOST", "DB_PORT", "DB_USER", "DB_PWD", "RABBITMQ_USER", "RABBITMQ_PWD"}

func parseEnv() {
	for _, v := range variableLists {
		if os.Getenv(v) == "" {
			log.Fatalf("%s is empty!", v)
		}
		envs[v] = os.Getenv(v)
	}
}

func startService() {

	//err := handlers.InitDatabaseCon(envs["DB_HOST"], envs["DB_PORT"], envs["DB_USER"], envs["DB_PWD"], "mydb")

	err := handlers.InitDatabaseCon(envs["DB_HOST"], envs["DB_PORT"], envs["DB_USER"], envs["DB_PWD"], "voice_emotion")
	if err != nil {
		log.Fatal("Can't connect to database!!!")
	}

	port, err := strconv.Atoi(envs["RABBITMQ_PORT"])
	if err != nil {
		log.Fatalf("Can't conver  RABBITMQ_PORT(%s) to int!!!", envs["RABBITMQ_PORT"])
	}

	handlers.StartReceiveTaskService(envs["RABBITMQ_HOST"], port, envs["RABBITMQ_USER"], envs["RABBITMQ_PWD"], handlers.QUEUEMAP["resultQueue"].Name, recordData)

}

func parseJSON(task *string) (*handlers.EmotionBlock, error) {
	eb := new(handlers.EmotionBlock) //emotion result

	//var eb handlers.EmotionBlock
	err := json.Unmarshal([]byte(*task), eb)
	if err != nil {
		return nil, err
	}

	return eb, nil
}

func recordData(task string) (string, string) {

	eb, err := parseJSON(&task)
	if err != nil {
		log.Println(err)
		return "", ""
	}

	idInt64, err := strconv.ParseUint(eb.ID, 10, 64)
	if err != nil {
		log.Println(err)
		return "", ""
	}
	eb.IDUint64 = idInt64

	// since render in ms (should voice-emotion worker modification)
	eb.RDuration /= 1000

	handlers.InsertAnalysisRecord(eb)
	handlers.ComputeChannelScore(eb)
	handlers.UpdateResult(eb)

	return "", ""
}

func fakeEnv() {
	envs["RABBITMQ_HOST"] = "127.0.0.1"
	envs["RABBITMQ_PORT"] = "5672"
	envs["DB_HOST"] = "127.0.0.1"
	envs["DB_PORT"] = "3306"
	envs["DB_USER"] = "root"
	envs["DB_PWD"] = "tyler"
	envs["FILE_PREFIX"] = "/Users/public/Documents"
	envs["LISTEN_PORT"] = ":8080"
	envs["RABBITMQ_USER"] = "root"
	envs["RABBITMQ_PWD"] = "tyler"
}

func main() {
	/*
		b1 := new(bytes.Buffer)
		f, _ := os.Open("test2.json")
		io.Copy(b1, f)
		f.Close()

		task := string(b1.Bytes())

		fakeEnv()
	*/

	parseEnv()
	//fakeEnv()
	handlers.InitDatabaseCon(envs["DB_HOST"], envs["DB_PORT"], envs["DB_USER"], envs["DB_PWD"], "voice_emotion")
	//handlers.QueryBlob()
	//recordData(task)

	startService()

}
