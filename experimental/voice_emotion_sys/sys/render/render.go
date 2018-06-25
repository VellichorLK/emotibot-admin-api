package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"handlers"

	"github.com/hashicorp/consul/api"
)

var envs = make(map[string]string)
var variableLists = [...]string{"RABBITMQ_HOST", "RABBITMQ_PORT", "DB_HOST", "DB_PORT", "DB_USER",
	"DB_PWD", "RABBITMQ_USER", "RABBITMQ_PWD", "CONSUL_IP", "CONSUL_PORT"}

var asrTaskQ = make(chan *handlers.EmotionBlock)

var asrQueName = "asr_tasks_queue"

//appid to timestamp
var timestamp = make(map[string]time.Time)
var cacheEnableAlert = make(map[string]string)
var cacheMailList = make(map[string][]string)
var cacheThreshold = make(map[string]map[int][]float64)
var consulClient *api.Client

func parseEnv() {
	for _, v := range variableLists {
		if os.Getenv(v) == "" {
			log.Fatalf("%s is empty!", v)
		}
		envs[v] = os.Getenv(v)
	}
}

func startService() {

	var err error
	// Get a new consul client
	config := api.DefaultConfig()
	config.Address = envs["CONSUL_IP"] + ":" + envs["CONSUL_PORT"]
	config.WaitTime = 1 * time.Second

	consulClient, err = api.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	//err := handlers.InitDatabaseCon(envs["DB_HOST"], envs["DB_PORT"], envs["DB_USER"], envs["DB_PWD"], "mydb")

	err = handlers.InitDatabaseCon(envs["DB_HOST"], envs["DB_PORT"], envs["DB_USER"], envs["DB_PWD"], "voice_emotion")
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

func getThreshold(appid string) (map[int][]float64, error) {
	querySQL := fmt.Sprintf("select %s,%s,%s from %s where %s=?",
		handlers.NCHANNEL, handlers.NType, handlers.NSCORE, handlers.ThresholdTable, handlers.NAPPID)

	rows, err := handlers.QuerySQL(querySQL, appid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	thresholdMap := make(map[int][]float64)

	emotionTypeMax := 2

	for rows.Next() {
		var ch, typ int
		var score float64
		var threshold []float64
		var ok bool
		err = rows.Scan(&ch, &typ, &score)
		if err != nil {
			return thresholdMap, err
		}
		if threshold, ok = thresholdMap[ch]; !ok {
			thresholdList := make([]float64, emotionTypeMax, emotionTypeMax)
			threshold = thresholdList
			thresholdMap[ch] = threshold
		}
		if typ < emotionTypeMax {
			threshold[typ] = score
		} else {
			log.Printf("Error: emotion type(%d) out of range(%d)\n", typ, emotionTypeMax)
		}

	}
	return thresholdMap, err

}

func getMailList(appid string) ([]string, error) {
	querySQL := fmt.Sprintf("select %s from %s where %s=?", handlers.NEmail, handlers.EmailTable, handlers.NAPPID)

	rows, err := handlers.QuerySQL(querySQL, appid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	mailList := make([]string, 0)
	for rows.Next() {
		var mail string
		err = rows.Scan(&mail)
		if err != nil {
			return mailList, err
		}
		mailList = append(mailList, mail)
	}

	return mailList, err
}

var notifySubject = "竹间语音情绪质检系统警告邮件。坐席号 %s"
var notifyBody = `竹间语音情绪质检系统分析录音发现愤怒情绪值超过 %v

录音文件名: %s
通话时长: %v
CallID: %v
客服坐席号: %s
通话时间: %s
上传时间: %s
客服愤怒情绪指数: %v
客户愤怒情绪指数: %v
`

func sendMail(subject string, body string, to []string) error {
	ms := &handlers.MailSender{Sender: "voice-admin@emotibot.com", Password: "Emotibot2018",
		MailServer: "mail.emotibot.com", MailServerPort: 25}
	return ms.SendMail(to, subject, body)
}

func notifyAlert(anger1 float64, anger2 float64, threshold float64, id uint64, appid string) error {
	sqlQuery := fmt.Sprintf("select %s, %s, %s, %s, %s, %s from %s where %s=?",
		handlers.NFILENAME, handlers.NRDURATION, handlers.NTAG, handlers.NTAG2, handlers.NFILET, handlers.NUPT,
		handlers.MainTable, handlers.NID)
	rows, err := handlers.QuerySQL(sqlQuery, id)
	if err != nil {
		return err
	}
	defer rows.Close()

	var fileName, staffID, callID string
	var duration int
	var createdTime, uploadTime int64
	if rows.Next() {
		err = rows.Scan(&fileName, &duration, &callID, &staffID, &createdTime, &uploadTime)
		if err != nil {
			log.Printf("Scan error: %v\n", err)
		} else {
			subject := fmt.Sprintf(notifySubject, staffID)
			duration = duration / 1000
			durationFormat := fmt.Sprintf("%d分%d秒", duration/60, duration%60)
			createT := time.Unix(createdTime, 0)
			uploadT := time.Unix(uploadTime, 0)
			body := fmt.Sprintf(notifyBody, threshold, fileName, durationFormat, callID, staffID,
				createT.Format(time.RFC1123Z), uploadT.Format(time.RFC1123Z), anger2, anger1)
			err = sendMail(subject, body, cacheMailList[appid])
			fmt.Printf("Send mail\n\nTo:%v\n Subject:%s\n%s\n", cacheMailList[appid], subject, body)
		}
	}
	return err
}

func checkAlert(scores []*handlers.TotalEmotionScore, appid string, id uint64) error {

	t, ok := timestamp[appid]
	if !ok {
		t = time.Unix(time.Now().Unix()-6, 0)
		timestamp[appid] = t
	}

	if time.Now().Unix()-t.Unix() > 5 {
		//polling consul
		timestamp[appid] = time.Now()
		consulKey := handlers.ConsulAlertKey + "/" + appid
		kv := consulClient.KV()

		// Lookup consul for updating enable alert or not
		pair, _, err := kv.Get(consulKey, nil)
		if err != nil {
			log.Printf("Error: consul error, %v, use cacheEnableAlert:%v\n", err, cacheEnableAlert[appid])
		} else if pair != nil {
			cacheEnableAlert[appid] = string(pair.Value)
		}

		//update the scores threshold
		threshold, err := getThreshold(appid)
		if err != nil {
			log.Printf("Error: getting threshold from db error %v\n", err)
		} else {
			cacheThreshold[appid] = threshold
		}

		//update the mail list
		mailList, err := getMailList(appid)
		if err != nil {
			log.Printf("Error: getting mail from db error %v\n", err)
		} else {
			cacheMailList[appid] = mailList
		}

	}

	goNotify := false
	var anger1, anger2 float64
	var thresholdV float64

	if cacheEnableAlert[appid] == "1" &&
		len(cacheThreshold[appid]) > 0 &&
		len(cacheMailList[appid]) > 0 {
		for _, s := range scores {
			if s.EType == 1 {
				if s.Channel == 1 {
					anger1 = s.Score
				} else {
					anger2 = s.Score
				}
			}

			if threshold, ok := cacheThreshold[appid][s.Channel]; ok {
				if s.EType <= len(threshold) && s.Score > threshold[s.EType] && threshold[s.EType] > 0 {
					//notice here, only allow one threshold value
					thresholdV = threshold[s.EType]
					goNotify = true
				}
			}
		}
	}

	if goNotify {
		return notifyAlert(anger1, anger2, thresholdV, id, appid)
	}

	return nil
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
	scores := handlers.ComputeChannelScore(eb)
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

	appid, err := handlers.GetAppidByID(eb.IDUint64)
	if err != nil || appid == "" {
		log.Printf("Error: can't find %v 's appid [%s]\n", eb.IDUint64, err)
	} else {
		err = checkAlert(scores, appid, eb.IDUint64)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}
	}

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
