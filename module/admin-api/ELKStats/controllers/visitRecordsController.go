package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
	"emotibot.com/emotigo/module/admin-api/ELKStats/services"
	"emotibot.com/emotigo/module/admin-api/util"
)

func VisitRecordsGetHandler(w http.ResponseWriter, r *http.Request) {
	enterpriseID := util.GetEnterpriseID(r)
	appID := util.GetAppID(r)

	if enterpriseID == "" && appID == "" {
		errResp := data.ErrorResponse{
			Message: fmt.Sprintf("Both headers %s and %s are not specified",
				data.EnterpriseIDHeaderKey, data.AppIDHeaderKey),
		}
		w.WriteHeader(http.StatusBadRequest)
		writeResponseJSON(w, errResp)
		return
	}

	VisitRecordsRequest := data.VisitRecordsRequest{}
	err := json.NewDecoder(r.Body).Decode(&VisitRecordsRequest)
	if err != nil {
		errResponse := data.NewErrorResponseWithCode(data.ErrCodeInvalidRequestBody,
			data.ErrInvalidRequestBody.Error())
		returnBadRequest(w, errResponse)
		return
	}

	startTime, endTime := util.CreateTimeRangeFromTimestamp(VisitRecordsRequest.StartTime,
		VisitRecordsRequest.EndTime)

	query := data.VisitRecordsQuery{
		CommonQuery: data.CommonQuery{
			EnterpriseID: enterpriseID,
			AppID:        appID,
			StartTime:    startTime,
			EndTime:      endTime,
		},
		Page: VisitRecordsRequest.Page,
	}

	if VisitRecordsRequest.Filter != nil {
		query.UserID = VisitRecordsRequest.Filter.UserID

		if VisitRecordsRequest.Filter.Contains != nil && VisitRecordsRequest.Filter.Contains.Type == "question" {
			query.Question = VisitRecordsRequest.Filter.Contains.Text
		}

		emotions := VisitRecordsRequest.Filter.Emotions
		if emotions != nil && len(emotions) > 0 && emotions[0].Type == "emotion" {
			group := emotions[0].Group
			if group != nil && len(group) > 0 {
				query.Emotion = group[0].Text
			}
		}

		qTypes := VisitRecordsRequest.Filter.QTypes
		if qTypes != nil && len(qTypes) > 0 && qTypes[0].Type == "qtype" {
			group := qTypes[0].Group
			if group != nil && len(group) > 0 {
				query.QType = group[0].ID
			}
		}
	}

	esCtx, esClient := util.GetElasticsearch()

	records, totalSize, limit, err := services.VisitRecordsQuery(esCtx, esClient, query)
	if err != nil {
		errResponse := data.NewErrorResponse(err.Error())
		returnInternalServerError(w, errResponse)
		return
	}

	response := data.VisitRecordsResponse{
		Data:        records,
		Limit:       limit,
		Page:        VisitRecordsRequest.Page,
		TableHeader: data.VisitRecordsTableHeader,
		TotalSize:   totalSize,
	}

	returnOK(w, response)
}
