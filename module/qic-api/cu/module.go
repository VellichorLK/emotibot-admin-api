package cu

import (
	"net/http"
	"strconv"
	"time"

	emotionengine "emotibot.com/emotigo/pkg/api/emotion-engine/v1"

	"emotibot.com/emotigo/pkg/logger"

	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/timecache"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo util.ModuleInfo
	// emotionTrain is an api instance that will be used in our api.
	emotionTrain func(apiModel emotionengine.Model) (modelID string, err error)
	// emotionPredict is an api instance that will be used in our api.
	emotionPredict func(request emotionengine.PredictRequest) (predictions []emotionengine.Predict, err error)
	// filterScore is the emotion filter standard
	filterScore = 60
	//KeyInitEmotionEngine is the key of one time function that will be used to trigger emotion-engine initial
	keyInitEmotionEngine = "Init emotion resource"
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "cu",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("POST", "text/process", []string{}, handleTextProcess),
			util.NewEntryPoint("POST", "conversation", []string{}, handleFlowCreate),
			util.NewEntryPoint("POST", "conversation/{id}/append", []string{}, handleFlowAdd),
			util.NewEntryPoint("GET", "conversation/{id}", []string{}, handleFlowResult),
			util.NewEntryPoint("PUT", "conversation/{id}", []string{}, handleFlowFinish),
		},
		OneTimeFunc: map[string]func(){
			"init db": func() {
				envs := ModuleInfo.Environments
				url := envs["MYSQL_URL"]
				user := envs["MYSQL_USER"]
				pass := envs["MYSQL_PASS"]
				db := envs["MYSQL_DB"]
				conn, err := util.InitDB(url, user, pass, db)
				if err != nil {
					logger.Error.Printf("Cannot init qi db, [%s:%s@%s:%s]: %s\n", user, pass, url, db, err.Error())
					return
				}
				util.SetDB(ModuleInfo.ModuleName, conn)
				serviceDao = model.NewSQLDao(conn)
				groupDao = model.NewGroupSQLDao(conn)
			},
			"init timecache": setUpTimeCache,
			keyInitEmotionEngine: func() {
				if ModuleInfo.Environments == nil {
					logger.Error.Println("Expect cu ModuleInfo.Environments is inited, but nil.")
					return
				}
				eeAddr, found := ModuleInfo.Environments["EMOTION_ENGINE_URL"]
				if !found {
					logger.Error.Println("Env EMOTION_ENGINE_URL is required!")
					return
				}
				client := emotionengine.Client{
					Transport: &http.Client{
						Timeout: time.Duration(3) * time.Second,
					},
					ServerURL: eeAddr,
				}
				emotionTrain = client.Train
				emotionPredict = client.Predict

				filterScoreText, found := ModuleInfo.Environments["EMOTION_FILTER_SCORE"]
				if found {
					score, err := strconv.Atoi(filterScoreText)
					if err != nil {
						logger.Error.Println("Variable EMOTION_FILTER_SCORE ", filterScoreText, " can not convert to int: ", err)
					}
					filterScore = score
				} else {
					logger.Warn.Println("Variable EMOTION_FILTER_SCORE is not found, use default value: ", filterScore)
				}
			},
		},
	}
}

func setUpTimeCache() {
	config := &timecache.TCacheConfig{}
	config.SetCollectionDuration(30 * time.Second)
	config.SetCollectionMethod(timecache.OnUpdate)
	cache.Activate(config)
}
