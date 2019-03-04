package qi

import (
	"net/http"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/pkg/logger"
)

func handleTrainAllTags(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	_, err := TrainModelByEnterprise(enterprise)
	if err != nil {
		if err == ErrTrainingBusy {
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusServiceUnavailable)
		} else {
			logger.Error.Printf("train failed. %s\n", err)
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		}
		return
	}
}

func handleUnload(w http.ResponseWriter, r *http.Request) {
	err := UnloadAllTags()
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
}

type modelHash struct {
	Models map[string][]modelResp `json:"models"`
}
type modelStatusResp struct {
	Models    []modelResp `json:"training"`
	OutOfDate bool        `json:"out_of_date"`
}
type modelResp struct {
	ID         int64 `json:"id,string"`
	CreateTime int64 `json:"create_time"`
	UpdateTime int64 `json:"update_time"`
}

var MStatWordings = map[int]string{
	MStatTraining:  "training",
	MStatReady:     "ready",
	MStatUsing:     "using",
	MStatErr:       "error",
	MStatDeprecate: "deprecate",
	MStatDeletion:  "deleted",
}

func handleTrainingStatus(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)

	models, err := GetModelByEnterprise(enterprise, MStatTraining)
	if err != nil {
		logger.Error.Printf("get models failed. %s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
	var resp modelStatusResp
	resp.Models = make([]modelResp, 0, len(models))
	for _, v := range models {
		m := modelResp{ID: int64(v.ID), CreateTime: v.CreateTime, UpdateTime: v.UpdateTime}
		resp.Models = append(resp.Models, m)
	}

	usingModels, err := GetModelByEnterprise(enterprise, MStatUsing)
	if err != nil {
		logger.Error.Printf("get models failed. %s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
	if len(usingModels) == 0 {
		resp.OutOfDate = true
	} else {
		//assume only one using model
		lastTime := usingModels[0].UpdateTime
		//because tagQuery uses >=
		tagQuery := model.TagQuery{UpdateTimeStart: lastTime + 1}
		tags, err := TagsByQuery(tagQuery)
		if err != nil {
			logger.Error.Printf("get tags failed. %s\n", err)
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
			return
		}
		if len(tags) != 0 {
			resp.OutOfDate = true
		}
	}

	err = util.WriteJSON(w, resp)
	if err != nil {
		logger.Error.Printf("write json failed. %s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.JSON_PARSE_ERROR, err.Error()), http.StatusInternalServerError)
	}
}

func handleTrainStatus(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)

	models, err := GetAllModelByEnterprise(enterprise)
	if err != nil {
		logger.Error.Printf("get models failed. %s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	var resp modelHash
	resp.Models = make(map[string][]modelResp)

	for _, v := range MStatWordings {
		resp.Models[v] = make([]modelResp, 0, 1)
	}

	for _, v := range models {
		m := modelResp{ID: int64(v.ID), CreateTime: v.CreateTime, UpdateTime: v.UpdateTime}
		key, ok := MStatWordings[v.Status]
		if !ok {
			logger.Warn.Printf("model %d has unknown status %d\n", m.ID, v.Status)
		} else {
			switch v.Status {
			case MStatUsing:
				fallthrough
			case MStatTraining:
				fallthrough
			case MStatReady:
				fallthrough
			case MStatErr:
				fallthrough
			case MStatDeprecate:
				fallthrough
			case MStatDeletion:
				resp.Models[key] = append(resp.Models[key], m)
			default:
				logger.Warn.Printf("model %d has unknown status %d\n", m.ID, v.Status)
			}
		}
	}

	err = util.WriteJSON(w, resp)
	if err != nil {
		logger.Error.Printf("write json failed. %s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.JSON_PARSE_ERROR, err.Error()), http.StatusInternalServerError)
	}
}
