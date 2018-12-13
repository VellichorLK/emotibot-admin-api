package cu

import (
	"net/http"
	"math/rand"
    "time"	
	"emotibot.com/emotigo/module/admin-api/util"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo  util.ModuleInfo
	maxDirDepth int
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "cu",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("POST", "text/process", []string{}, handleTextProcess),
		},
	}
	maxDirDepth = 4
}

func random(min, max int) int {
    rand.Seed(time.Now().Unix())
    return rand.Intn(max - min) + min
}

func handleTextProcess(w http.ResponseWriter, r *http.Request) {
	mockEmotions := []string{
		"不满",
		"称赞",
		"不喜欢",
		"高兴",
		"伤心",
		"害怕",
	}

	type RequestObj struct {
		Text string `json:"text"`
	}

	type EmotionObj struct {
		Label string `json:"label"`
	}
	type ResponseObj struct {
		Text string `json:"text"`
		Emotion []EmotionObj `json:"emotion"`
	}

	reqBody := []RequestObj{}
	err := util.ReadJSON(r, &reqBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}


	responseBody := []ResponseObj{}
	for _, textObj := range reqBody{
		ind := random(0,6)
		emotion := mockEmotions[ind]

		responseObj := ResponseObj{
			Text: textObj.Text,
			Emotion: []EmotionObj{
				EmotionObj{
					Label: emotion,
				},
			},
		}

		responseBody = append(responseBody, responseObj)
	}

	util.WriteJSON(w, responseBody)

}