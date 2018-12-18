package cu

import (
	"database/sql"
	"net/http"
	"time"

	emotionengine "emotibot.com/emotigo/pkg/api/emotion-engine/v1"

	"emotibot.com/emotigo/pkg/logger"

	"emotibot.com/emotigo/module/admin-api/util"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo util.ModuleInfo
	// emotionTrain is an api instance that will be used in our api.
	emotionTrain func(apiModel emotionengine.Model) (modelID string, err error)
	// emotionPredict is an api instance that will be used in our api.
	emotionPredict func(request emotionengine.PredictRequest) (predictions []emotionengine.Predict, err error)
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "cu",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("POST", "text/process", []string{}, handleTextProcess),
			util.NewEntryPoint("POST", "conversation", []string{}, handleFlowCreate),
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
