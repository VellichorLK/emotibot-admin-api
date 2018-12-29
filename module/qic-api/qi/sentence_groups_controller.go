package qi

import (
	autil "emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/pkg/logger"
	"net/http"
)

var roleMapping map[string]int = map[string]int{
	"staff":    0,
	"customer": 1,
}

var positionMap map[string]int = map[string]int{
	"top":    0,
	"bottom": 1,
}

func sentenceGroupInReqToSentenceGroup(sentenceGroupInReq *SentenceGroupInReq) (group *model.SentenceGroup) {
	group = &model.SentenceGroup{
		Name: sentenceGroupInReq.Name,
	}

	sentences := []model.SimpleSentence{}
	for _, sid := range sentenceGroupInReq.Sentences {
		sentence := model.SimpleSentence{
			UUID: sid,
		}
		sentences = append(sentences, sentence)
	}
	group.Sentences = sentences

	if roleCode, ok := roleMapping[sentenceGroupInReq.Role]; ok {
		group.Role = roleCode
	} else {
		group.Role = -1
	}

	if poisitionCode, ok := positionMap[sentenceGroupInReq.Position]; ok {
		group.Position = poisitionCode
	} else {
		group.Position = -1
	}

	group.Distance = sentenceGroupInReq.PositionDistance
	return
}

func handleCreateSentenceGroup(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)

	groupInReq := SentenceGroupInReq{}
	err := autil.ReadJSON(r, &groupInReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	group := sentenceGroupInReqToSentenceGroup(&groupInReq)
	group.Enterprise = enterprise
	if group.Position == -1 || group.Role == -1 {
		http.Error(w, "bad sentence group", http.StatusBadRequest)
		return
	}

	createdGroup, err := CreateSentenceGroup(group)
	if err != nil {
		logger.Error.Printf("error while create sentence in handleCreateSentenceGroup, reason: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	groupInResponse := SentenceGroupInResponse{
		ID: createdGroup.UUID,
	}
	autil.WriteJSON(w, groupInResponse)
	return
}

func handleGetSentenceGroups(w http.ResponseWriter, r *http.Request) {

}

func handleGetSentenceGroup(w http.ResponseWriter, r *http.Request) {

}

func handleUpdateSentenceGroup(w http.ResponseWriter, r *http.Request) {

}

func handleDeleteSentenceGroup(w http.ResponseWriter, r *http.Request) {

}
