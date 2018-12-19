package cu

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	emotionengine "emotibot.com/emotigo/pkg/api/emotion-engine/v1"

	"emotibot.com/emotigo/pkg/logger"

	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/qic-api/util/timecache"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo util.ModuleInfo
	// emotionTrain is an api instance that will be used in our api.
	emotionTrain func(apiModel emotionengine.Model) (modelID string, err error)
	// emotionPredict is an api instance that will be used in our api.
	emotionPredict func(request emotionengine.PredictRequest) (predictions []emotionengine.Predict, err error)
	filterScore    = 60
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "cu",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("POST", "text/process", []string{}, handleTextProcess),
			util.NewEntryPoint("POST", "conversation", []string{}, handleFlowCreate),
			util.NewEntryPoint("POST", "conversation/{id}/append", []string{}, handleFlowAdd),
		},
		OneTimeFunc: map[string]func(){
			"Init emotion resource": func() {
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
				score, err := strconv.Atoi(filterScoreText)
				if err != nil {
					logger.Error.Println("Variable EMOTION_FILTER_SCORE ", filterScoreText, " can not convert to int: ", err)
				} else if found {
					filterScore = score
				} else {
					logger.Warn.Println("Variable EMOTION_FILTER_SCORE is not found, use default value: ", filterScore)
				}
			},
		},
	}
}

//SetupServiceDB sets up the db structure
func SetupServiceDB(db *sql.DB) {
	serviceDao = SQLDao{
		conn: db,
	}
}

func SetUpTimeCache() {
	config := &timecache.TCacheConfig{}
	config.SetCollectionDuration(30 * time.Second)
	config.SetCollectionMethod(timecache.OnUpdate)
	cache.Activate(config)
}
