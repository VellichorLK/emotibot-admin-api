package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"

	"handlers"
)

var envs = make(map[string]string)
var variableLists = [...]string{"RABBITMQ_HOST", "RABBITMQ_PORT", "DB_HOST", "DB_PORT", "DB_USER", "DB_PWD", "RABBITMQ_USER", "RABBITMQ_PWD"}

var asrTaskQ = make(chan *handlers.EmotionBlock)

var asrQueName = "asr_tasks_queue"

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

	queue, ok := handlers.QUEUEMAP["asrTaskQueue"]
	if ok {
		asrQueName = queue.Name
	}

	go handlers.StartSendTaskService(envs["RABBITMQ_HOST"], port, envs["RABBITMQ_USER"], envs["RABBITMQ_PWD"], taskReceive)

	handlers.StartReceiveTaskService(envs["RABBITMQ_HOST"], port, envs["RABBITMQ_USER"], envs["RABBITMQ_PWD"], handlers.QUEUEMAP["resultQueue"].Name, recordData)

}

//taskReceive return finished channel,task as string, fileID, queueName, priority
func taskReceive() (string, string, string, uint8) {

	for {
		taskInfo := <-asrTaskQ

		b, err := json.Marshal(taskInfo)
		if err != nil {
			log.Println(err)
			log.Printf("skip one task %v\n", taskInfo)
			continue
		}
		return string(b), taskInfo.ID, asrQueName, 0
	}

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
	startTime := time.Now()
	log.Println(" >>> recordData start, eb.ID = ", eb.ID, ", startTime   = ", startTime)

	idInt64, err := strconv.ParseUint(eb.ID, 10, 64)
	if err != nil {
		log.Println(err)
		return "", ""
	}
	eb.IDUint64 = idInt64

	err = handlers.InsertAnalysisRecord(eb)
	if err != nil {
		log.Println(err)
		return "", ""
	}
	handlers.ComputeChannelScore(eb)
	handlers.UpdateResult(eb)

	var asrPush bool

	asrTaskQ <- eb
	asrPush = <-handlers.RelyQueue

	//max retry
	for i := 0; i < 3 && !asrPush; i++ {
		time.Sleep(2 * time.Second)
		asrTaskQ <- eb
		asrPush = <-handlers.RelyQueue
	}

	if true == getEnvAsBoolean("ENABLE_SILENCE_COMPUTING", false) {
		handlers.ComputeSilence(eb)
	}

	log.Println(">>> recordData   end, eb.ID = ", eb.ID, ", elapsedTime = ", time.Now().Sub(startTime))
	return "", ""
}

func getEnvAsBoolean(keyName string, defaultValue bool) bool {
	keyValue, keyExist := os.LookupEnv(keyName)
	if false == keyExist {
		return defaultValue
	}
	parseResult, err := strconv.ParseBool(keyValue)
	if err != nil {
		return defaultValue
	}
	return parseResult
}

func fakeEnv() {
	envs["RABBITMQ_HOST"] = "127.0.0.1"
	envs["RABBITMQ_PORT"] = "5672"
	envs["DB_HOST"] = "127.0.0.1"
	envs["DB_PORT"] = "3306"
	envs["DB_USER"] = "root"
	envs["DB_PWD"] = "password"
	envs["FILE_PREFIX"] = "/Users/public/Documents"
	envs["LISTEN_PORT"] = ":8080"
	envs["RABBITMQ_USER"] = "guest"
	envs["RABBITMQ_PWD"] = "guest"
}

func main() {

	parseEnv()
	//fakeEnv()
	handlers.InitDatabaseCon(envs["DB_HOST"], envs["DB_PORT"], envs["DB_USER"], envs["DB_PWD"], "voice_emotion")
	//handlers.QueryBlob()
	//recordData(task)

	startService()

}
