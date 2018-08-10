package controllers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
	"emotibot.com/emotigo/module/admin-api/ELKStats/services"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/elasticsearch"
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

	visitRecordsRequest := data.VisitRecordsRequest{}
	err := json.NewDecoder(r.Body).Decode(&visitRecordsRequest)
	if err != nil {
		errResponse := data.NewErrorResponseWithCode(data.ErrCodeInvalidRequestBody,
			data.ErrInvalidRequestBody.Error())
		returnBadRequest(w, errResponse)
		return
	}

	startTime := time.Unix(visitRecordsRequest.StartTime, 0).Local()
	endTime := time.Unix(visitRecordsRequest.EndTime, 0).Local()

	query := data.VisitRecordsQuery{
		CommonQuery: data.CommonQuery{
			EnterpriseID: enterpriseID,
			AppID:        appID,
			StartTime:    startTime,
			EndTime:      endTime,
		},
		Page: visitRecordsRequest.Page,
	}

	if visitRecordsRequest.Filter != nil {
		query.UserID = visitRecordsRequest.Filter.UserID

		if visitRecordsRequest.Filter.Contains != nil && visitRecordsRequest.Filter.Contains.Type == "question" {
			query.Question = visitRecordsRequest.Filter.Contains.Text
		}

		emotions := visitRecordsRequest.Filter.Emotions
		if emotions != nil && len(emotions) > 0 && emotions[0].Type == "emotion" {
			query.Emotions = make([]interface{}, 0)
			for _, group := range emotions[0].Group {
				query.Emotions = append(query.Emotions, group.Text)
			}
		}

		qTypes := visitRecordsRequest.Filter.QTypes
		if qTypes != nil && len(qTypes) > 0 && qTypes[0].Type == "qtype" {
			group := qTypes[0].Group
			if group != nil && len(group) > 0 {
				query.QType = group[0].ID
			}
		}

		if visitRecordsRequest.Filter.Tags != nil {
			query.Tags = make([]data.VisitRecordsQueryTag, 0)

			for _, queryTag := range visitRecordsRequest.Filter.Tags {
				if queryTag.Group != nil && len(queryTag.Group) > 0 {
					tags := make([]string, 0)
					for _, t := range queryTag.Group {
						tags = append(tags, t.Text)
					}

					tag := data.VisitRecordsQueryTag{
						Type:  queryTag.Type,
						Texts: tags,
					}
					query.Tags = append(query.Tags, tag)
				}
			}
		}
	}

	esCtx, esClient := elasticsearch.GetClient()
	if esClient == nil {
		returnInternalServerError(w, data.NewErrorResponse(data.ErrNotInit.Error()))
		return
	}

	if !visitRecordsRequest.Export {
		if visitRecordsRequest.Limit != 0 {
			query.PageLimit = visitRecordsRequest.Limit
		} else {
			query.PageLimit = data.VisitRecordsPageLimit
		}

		records, totalSize, limit, err := services.VisitRecordsQuery(esCtx, esClient, query)
		if err != nil {
			errResponse := data.NewErrorResponse(err.Error())
			returnInternalServerError(w, errResponse)
			return
		}

		response := data.VisitRecordsResponse{
			Data:        records,
			Limit:       limit,
			Page:        visitRecordsRequest.Page,
			TableHeader: data.VisitRecordsTableHeader,
			TotalSize:   totalSize,
		}

		returnOK(w, response)
	} else {
		query.PageLimit = data.VisitRecordsExportPageLimit
		query.Page = 1

		totalRecords := make([]*data.VisitRecordsData, 0)

		records, totalSize, _, err := services.VisitRecordsQuery(esCtx, esClient, query)
		if err != nil {
			errResponse := data.NewErrorResponse(err.Error())
			returnInternalServerError(w, errResponse)
			return
		}

		totalRecords = append(totalRecords, records...)

		for int64(query.PageLimit*query.Page) < totalSize {
			query.Page++
			records, _, _, err := services.VisitRecordsQuery(esCtx, esClient, query)
			if err != nil {
				errResponse := data.NewErrorResponse(err.Error())
				returnInternalServerError(w, errResponse)
				return
			}

			totalRecords = append(totalRecords, records...)
		}

		if len(totalRecords) == 0 {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Everything goes right, start writting CSV response
		out := csv.NewWriter(w)
		out.Write(data.VisitRecordsExportHeader)

		for _, record := range totalRecords {
			csvData := []string{
				record.UserID,
				record.UserQ,
				strconv.FormatFloat(record.Score, 'f', -1, 64),
				record.StdQ,
				record.Answer,
				record.LogTime,
				record.Emotion,
				record.QType,
			}

			err = out.Write(csvData)
			if err != nil {
				errResponse := data.NewErrorResponse(err.Error())
				returnInternalServerError(w, errResponse)
				return
			}
		}

		out.Flush()
		err = out.Error()
		if err != nil {
			errResponse := data.NewErrorResponse(err.Error())
			returnInternalServerError(w, errResponse)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
