package v2

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	intentengine "emotibot.com/emotigo/module/admin-api/intentengine/v2"

	"emotibot.com/emotigo/module/admin-api/ELKStats/controllers"
	"emotibot.com/emotigo/module/admin-api/ELKStats/controllers/common"
	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
	dataCommon "emotibot.com/emotigo/module/admin-api/ELKStats/data/common"
	dataV2 "emotibot.com/emotigo/module/admin-api/ELKStats/data/v2"
	"emotibot.com/emotigo/module/admin-api/ELKStats/services"
	servicesCommon "emotibot.com/emotigo/module/admin-api/ELKStats/services/common"
	servicesV2 "emotibot.com/emotigo/module/admin-api/ELKStats/services/v2"
	"emotibot.com/emotigo/module/admin-api/util/elasticsearch"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	dal "emotibot.com/emotigo/pkg/api/dal/v1"
	"emotibot.com/emotigo/pkg/logger"
	"github.com/gorilla/mux"
	dac "emotibot.com/emotigo/pkg/api/dac/v1"
)

//VisitRecordsGetHandler handle advanced query for records.
//Limit & Page should be given by r's query string.
func VisitRecordsGetHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	query, err, errCode := newRecordQuery(r)
	if err != nil {
		errResponse := data.NewErrorResponseWithCode(errCode, err.Error())
		controllers.ReturnBadRequest(w, errResponse)
		return
	}

	if query.StartTime.After(query.EndTime) {
		errResponse := data.NewErrorResponseWithCode(data.ErrCodeInvalidParameterStartTime,
			"start_time cannot greater than end_time")
		controllers.ReturnBadRequest(w, errResponse)
		return
	}

	result, err := servicesV2.VisitRecordsQuery(query)
	if err != nil {
		var errResponse data.ErrorResponse
		if rootCauseErrors, ok := elasticsearch.ExtractElasticsearchRootCauseErrors(err); ok {
			errResponse = data.NewErrorResponse(strings.Join(rootCauseErrors, ", "))
		} else {
			errResponse = data.NewErrorResponse(err.Error())
		}
		controllers.ReturnInternalServerError(w, errResponse)
		return
	}

	locale := requestheader.GetLocale(r)
	response := dataV2.VisitRecordsResponse{
		TableHeader: dataV2.VisitRecordsTableHeader[locale],
		Data:        result.Data,
		TotalSize:   result.TotalSize,
		IgnoredSize: result.IgnoredSize,
		MarkedSize:  result.MarkedSize,
		Page:        query.From/query.Limit + 1,
		Limit:       query.Limit,
	}

	controllers.ReturnOK(w, response)
}

func VisitRecordsExportHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	query, err, errCode := newRecordQuery(r)
	if err != nil {
		errResponse := data.NewErrorResponseWithCode(errCode, err.Error())
		controllers.ReturnBadRequest(w, errResponse)
		return
	}

	if query.StartTime.After(query.EndTime) {
		errResponse := data.NewErrorResponseWithCode(data.ErrCodeInvalidParameterStartTime,
			"start_time cannot greater than end_time")
		controllers.ReturnBadRequest(w, errResponse)
		return
	}

	locale := requestheader.GetLocale(r)
	exportTaskID, err := servicesV2.VisitRecordsExport(query, locale)
	if err != nil {
		var errResponse data.ErrorResponse
		if rootCauseErrors, ok := elasticsearch.ExtractElasticsearchRootCauseErrors(err); ok {
			errResponse = data.NewErrorResponse(strings.Join(rootCauseErrors, "; "))
		} else {
			errResponse = data.NewErrorResponse(err.Error())

			switch err {
			case data.ErrExportTaskInProcess:
				controllers.ReturnForbiddenRequest(w, errResponse)
			default:
				controllers.ReturnInternalServerError(w, errResponse)
			}
		}
		return
	}

	response := dataV2.VisitRecordsExportResponse{
		ExportID: exportTaskID,
	}

	controllers.ReturnOK(w, response)
}

func VisitRecordsExportDownloadHandler(w http.ResponseWriter, r *http.Request) {
	common.VisitRecordsExportDownloadHandler(w, r)
}

func VisitRecordsExportDeleteHandler(w http.ResponseWriter, r *http.Request) {
	common.VisitRecordsExportDeleteHandler(w, r)
}

func VisitRecordsExportStatusHandler(w http.ResponseWriter, r *http.Request) {
	common.VisitRecordsExportStatusHandler(w, r)
}

// RecordsIgnoredUpdateHandler handle the request for updating record's ignore column
func RecordsIgnoredUpdateHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Ignore  *bool
		Records []string `json:"records"`
	}
	defer r.Body.Close()

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		controllers.ReturnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, err.Error()))
		return
	}

	if request.Ignore == nil {
		controllers.ReturnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, "ignore is required"))
		return
	}

	if request.Records == nil {
		controllers.ReturnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, "records is required"))
		return
	}

	q := &dataV2.VisitRecordsQuery{
		CommonQuery: data.CommonQuery{
			AppID: requestheader.GetAppID(r),
		},
		RecordIDs: make([]interface{}, len(request.Records)),
	}

	for i, r := range request.Records {
		q.RecordIDs[i] = r
	}

	err = servicesV2.UpdateRecords(q, servicesV2.UpdateRecordIgnore(*request.Ignore))
	if err != nil {
		if rootCauseErrors, ok := elasticsearch.ExtractElasticsearchRootCauseErrors(err); ok {
			errResponse := data.NewErrorResponse(strings.Join(rootCauseErrors, "; "))
			controllers.ReturnInternalServerError(w, errResponse)
		} else {
			controllers.ReturnBadRequest(w, data.NewErrorResponseWithCode(http.StatusInternalServerError,
				"update records failed, "+err.Error()))
		}

		return
	}

	w.WriteHeader(http.StatusOK)
}

// NewRecordsMarkUpdateHandler create a handler to handle records mark & unmark request.
// The handler will update record store & ssm store.
// Because we separate the Init() function and Controller, the only way to pass dal.Client is from parameters.
func NewRecordsMarkUpdateHandler(client *dal.Client) func(w http.ResponseWriter, r *http.Request) {
	if client == nil {
		return func(w http.ResponseWriter, r *http.Request) {
			controllers.ReturnInternalServerError(w, data.NewErrorResponse("no valid dal client"))
		}
	}

	type internalError struct {
		IsRollbacked bool `json:"rollbacked"`
	}

	type response struct {
		Done []interface{} `json:"done"`
		Skip []string      `json:"skip"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Content string   `json:"content"`
			Mark    *bool    `json:"mark"`
			Records []string `json:"records"`
		}
		defer r.Body.Close()
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			controllers.ReturnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, err.Error()))
			return
		}
		if request.Mark == nil {
			controllers.ReturnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, "mark is required."))
			return
		}
		if request.Records == nil {
			controllers.ReturnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, "records is required"))
			return
		}

		appID := requestheader.GetAppID(r)
		q := &dataV2.VisitRecordsQuery{
			CommonQuery: data.CommonQuery{
				AppID: appID,
			},
			RecordIDs: make([]interface{}, len(request.Records)),
			Limit:     int64(len(request.Records)),
		}

		for i, r := range request.Records {
			q.RecordIDs[i] = r
		}

		// Need retrive record's userq
		result, err := servicesV2.VisitRecordsQuery(q)
		if err != nil {
			logger.Error.Printf("check record content failed, %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(internalError{
				IsRollbacked: true,
			})
			return
		}

		todoRecords := make([]*dataV2.VisitRecordsData, 0)
		skipRecordsID := make([]string, 0)

		// Filter skip records
		for _, record := range result.Data {
			var isSQ bool
			isSQ, err = client.IsStandardQuestion(appID, record.UserQ)
			if err != nil {
				logger.Error.Printf("Get isStd failed, %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(internalError{
					IsRollbacked: true,
				})
				return
			}

			if isSQ {
				skipRecordsID = append(skipRecordsID, record.UniqueID)
				continue
			}
			todoRecords = append(todoRecords, record)
		}

		// Deduplicate input userQ, because dalClient can not handle duplicate request.
		questions := map[string]struct{}{}
		for _, r := range todoRecords {
			_, found := questions[r.UserQ]
			if found {
				continue
			}
			questions[r.UserQ] = struct{}{}
		}
		uniqueUserQ := []string{}
		for r := range questions {
			uniqueUserQ = append(uniqueUserQ, r)
		}

		if *request.Mark {
			// If delete Simq have problem, we will know at next step setSimQ, so dont bother to check delete simQ operation at all.
			client.DeleteSimilarQuestions(appID, uniqueUserQ...)
			err = client.SetSimilarQuestion(appID, request.Content, uniqueUserQ...)
			if err != nil {
				logger.Error.Printf("set simQ failed, %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(internalError{
					IsRollbacked: false,
				})
				return
			}
		} else {
			// Unmark should remove the records from ssm store
			err = client.DeleteSimilarQuestions(appID, uniqueUserQ...)
			// Note: because ssm wont sync with record, so it can be someone delete the ssm
			// We only need to return error if the problem is not "not exist"
			if dErr, ok := err.(*dal.DetailError); ok {
				for _, op := range dErr.Results {
					if op != "NOT_EXIST" {
						logger.Error.Printf("delete sim q failed, %v", err)
						w.WriteHeader(http.StatusInternalServerError)
						json.NewEncoder(w).Encode(internalError{
							IsRollbacked: false,
						})
						return
					}
				}
			} else if err != nil {
				logger.Error.Printf("delete sim q failed, %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(internalError{
					IsRollbacked: false,
				})
				return
			}
		}

		newQuery := &dataV2.VisitRecordsQuery{
			CommonQuery: data.CommonQuery{
				AppID: appID,
			},
			RecordIDs: make([]interface{}, len(todoRecords)),
		}

		for i, r := range todoRecords {
			newQuery.RecordIDs[i] = r.UniqueID
		}

		err = servicesV2.UpdateRecords(newQuery, servicesCommon.UpdateRecordMark(*request.Mark))
		if err != nil {
			logger.Error.Printf("service update record failed, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(internalError{
				IsRollbacked: true,
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		resp := response{
			Done: newQuery.RecordIDs,
			Skip: skipRecordsID,
		}
		data, _ := json.Marshal(resp)
		w.Write(data)
	}
}


// NewRecordsMarkUpdateHandler create a handler to handle records mark & unmark request.
// The handler will update record store & ssm store.
// Because we separate the Init() function and Controller, the only way to pass dac.Client is from parameters.
func NewRecordsMarkUpdateHandlerV3(client *dac.Client) func(w http.ResponseWriter, r *http.Request) {
	if client == nil {
		return func(w http.ResponseWriter, r *http.Request) {
			controllers.ReturnInternalServerError(w, data.NewErrorResponse("no valid dac client"))
		}
	}

	type internalError struct {
		IsRollbacked bool `json:"rollbacked"`
	}

	type response struct {
		Done []interface{} `json:"done"`
		Skip []string      `json:"skip"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Content string   `json:"content"`
			Mark    *bool    `json:"mark"`
			Records []string `json:"records"`
		}
		defer r.Body.Close()
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			controllers.ReturnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, err.Error()))
			return
		}
		if request.Mark == nil {
			controllers.ReturnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, "mark is required."))
			return
		}
		if request.Records == nil {
			controllers.ReturnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, "records is required"))
			return
		}

		appID := requestheader.GetAppID(r)
		q := &dataV2.VisitRecordsQuery{
			CommonQuery: data.CommonQuery{
				AppID: appID,
			},
			RecordIDs: make([]interface{}, len(request.Records)),
			Limit:     int64(len(request.Records)),
		}

		for i, r := range request.Records {
			q.RecordIDs[i] = r
		}

		// Need retrive record's userq
		result, err := servicesV2.VisitRecordsQuery(q)
		if err != nil {
			logger.Error.Printf("check record content failed, %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(internalError{
				IsRollbacked: true,
			})
			return
		}

		todoRecords := make([]*dataV2.VisitRecordsData, 0)
		skipRecordsID := make([]string, 0)

		// Filter skip records
		for _, record := range result.Data {
			var isSQ bool
			isSQ, err = client.IsStandardQuestion(appID, record.UserQ)
			if err != nil {
				logger.Error.Printf("Get isStd failed, %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(internalError{
					IsRollbacked: true,
				})
				return
			}

			if isSQ {
				skipRecordsID = append(skipRecordsID, record.UniqueID)
				continue
			}
			todoRecords = append(todoRecords, record)
		}

		// Deduplicate input userQ, because dacClient can not handle duplicate request.
		questions := map[string]struct{}{}
		for _, r := range todoRecords {
			_, found := questions[r.UserQ]
			if found {
				continue
			}
			questions[r.UserQ] = struct{}{}
		}
		uniqueUserQ := []string{}
		for r := range questions {
			uniqueUserQ = append(uniqueUserQ, r)
		}

		if *request.Mark {
			// If delete Simq have problem, we will know at next step setSimQ, so dont bother to check delete simQ operation at all.
			client.DeleteSimilarQuestions(appID, uniqueUserQ...)
			err = client.SetSimilarQuestion(appID, request.Content, uniqueUserQ...)
			if err != nil {
				logger.Error.Printf("set simQ failed, %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(internalError{
					IsRollbacked: false,
				})
				return
			}
		} else {
			// Unmark should remove the records from ssm store
			err = client.DeleteSimilarQuestions(appID, uniqueUserQ...)
			// Note: because ssm wont sync with record, so it can be someone delete the ssm
			// We only need to return error if the problem is not "not exist"
			if dErr, ok := err.(*dac.DetailError); ok {
				for _, op := range dErr.Results {
					if op != "NOT_EXIST" {
						logger.Error.Printf("delete sim q failed, %v", err)
						w.WriteHeader(http.StatusInternalServerError)
						json.NewEncoder(w).Encode(internalError{
							IsRollbacked: false,
						})
						return
					}
				}
			} else if err != nil {
				logger.Error.Printf("delete sim q failed, %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(internalError{
					IsRollbacked: false,
				})
				return
			}
		}

		newQuery := &dataV2.VisitRecordsQuery{
			CommonQuery: data.CommonQuery{
				AppID: appID,
			},
			RecordIDs: make([]interface{}, len(todoRecords)),
		}

		for i, r := range todoRecords {
			newQuery.RecordIDs[i] = r.UniqueID
		}

		err = servicesV2.UpdateRecords(newQuery, servicesCommon.UpdateRecordMark(*request.Mark))
		if err != nil {
			logger.Error.Printf("service update record failed, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(internalError{
				IsRollbacked: true,
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		resp := response{
			Done: newQuery.RecordIDs,
			Skip: skipRecordsID,
		}
		data, _ := json.Marshal(resp)
		w.Write(data)
	}
}


// NewRecordSSMHandler create a handler for retriving ssm info of certain record.
func NewRecordSSMHandler(client *dal.Client) func(http.ResponseWriter, *http.Request) {
	if client == nil {
		return func(w http.ResponseWriter, r *http.Request) {
			controllers.ReturnInternalServerError(w, data.NewErrorResponse("no valid dal client"))
		}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := mux.Vars(r)["id"]
		if !ok {
			controllers.ReturnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, "no path variable id"))
			return
		}

		appID := requestheader.GetAppID(r)
		query := &dataV2.VisitRecordsQuery{
			CommonQuery: data.CommonQuery{
				AppID: appID,
			},
			RecordIDs: []interface{}{id},
			Limit:     1,
		}

		result, err := servicesV2.VisitRecordsQuery(query)
		if err != nil {
			logger.Error.Printf("fetch es records failed, %v\n", err)
			controllers.ReturnInternalServerError(w, data.NewErrorResponse("internal server error"))
			return
		}
		if size := len(result.Data); size == 0 || size > 1 {
			controllers.ReturnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, "record id "+id+" is ambiguous(results: "+strconv.Itoa(size)+")"))
			return
		}
		if !result.Data[0].IsMarked {
			controllers.ReturnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, "record is not marked"))
		}
		lq := result.Data[0].UserQ
		var response = struct {
			Content string `json:"marked_content"`
		}{}
		isLQ, err := client.IsSimilarQuestion(appID, lq)
		if err != nil {
			logger.Error.Printf("get isSimilarQ failed, %v", err)
			controllers.ReturnInternalServerError(w, data.NewErrorResponse("internal server error"))
			return
		}
		if !isLQ {
			response.Content = ""
			controllers.ReturnOK(w, response)
			return
		}

		response.Content, err = client.Question(appID, lq)
		if err != nil {
			logger.Error.Printf("get question for lq %s failed, %v", lq, err)
			controllers.ReturnInternalServerError(w, data.NewErrorResponse("internal server error"))
			return
		}
		controllers.ReturnOK(w, response)
	}
}



// NewRecordSSMHandler create a handler for retriving ssm info of certain record.
func NewRecordSSMHandlerV3(client *dac.Client) func(http.ResponseWriter, *http.Request) {
	if client == nil {
		return func(w http.ResponseWriter, r *http.Request) {
			controllers.ReturnInternalServerError(w, data.NewErrorResponse("no valid dac client"))
		}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := mux.Vars(r)["id"]
		if !ok {
			controllers.ReturnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, "no path variable id"))
			return
		}

		appID := requestheader.GetAppID(r)
		query := &dataV2.VisitRecordsQuery{
			CommonQuery: data.CommonQuery{
				AppID: appID,
			},
			RecordIDs: []interface{}{id},
			Limit:     1,
		}

		result, err := servicesV2.VisitRecordsQuery(query)
		if err != nil {
			logger.Error.Printf("fetch es records failed, %v\n", err)
			controllers.ReturnInternalServerError(w, data.NewErrorResponse("internal server error"))
			return
		}
		if size := len(result.Data); size == 0 || size > 1 {
			controllers.ReturnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, "record id "+id+" is ambiguous(results: "+strconv.Itoa(size)+")"))
			return
		}
		if !result.Data[0].IsMarked {
			controllers.ReturnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, "record is not marked"))
		}
		lq := result.Data[0].UserQ
		var response = struct {
			Content string `json:"marked_content"`
		}{}
		isLQ, err := client.IsSimilarQuestion(appID, lq)
		if err != nil {
			logger.Error.Printf("get isSimilarQ failed, %v", err)
			controllers.ReturnInternalServerError(w, data.NewErrorResponse("internal server error"))
			return
		}
		if !isLQ {
			response.Content = ""
			controllers.ReturnOK(w, response)
			return
		}

		response.Content, err = client.Question(appID, lq)
		if err != nil {
			logger.Error.Printf("get question for lq %s failed, %v", lq, err)
			controllers.ReturnInternalServerError(w, data.NewErrorResponse("internal server error"))
			return
		}
		controllers.ReturnOK(w, response)
	}
}

func RecordIntentHandler(w http.ResponseWriter, r *http.Request) {
	id, ok := mux.Vars(r)["id"]
	if !ok {
		controllers.ReturnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, "no path variable id"))
		return
	}

	appid := requestheader.GetAppID(r)
	query := &dataV2.VisitRecordsQuery{
		CommonQuery: data.CommonQuery{
			AppID: appid,
		},
		RecordIDs: []interface{}{id},
		Limit:     1,
	}

	result, err := servicesV2.VisitRecordsQuery(query)
	if err != nil {
		logger.Error.Printf("fetch es records failed, %v\n", err)
		controllers.ReturnInternalServerError(w, data.NewErrorResponse("internal server error"))
		return
	}
	if size := len(result.Data); size == 0 || size > 1 {
		controllers.ReturnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, "record id "+id+" is ambiguous(results: "+strconv.Itoa(size)+")"))
		return
	}

	userQ := result.Data[0].UserQ
	intentID := result.Data[0].MarkedIntent
	if intentID == nil {
		controllers.ReturnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, "no marked intent"))
		return
	}
	origIntent, err := intentengine.GetIntent(appid, *intentID, "")
	if err != nil || origIntent == nil {
		controllers.ReturnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, "cannot get marked intent"))
		return
	}
	intent, err := intentengine.GetCurrentViaName(appid, nil, origIntent.Name)
	if err != nil || intent == nil {
		controllers.ReturnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, "marked intent is not existed in current version"))
		return
	}
	positive := true
	if intent.Negative != nil {
		for _, sentence := range *intent.Negative {
			if sentence == nil {
				continue
			}
			if userQ == sentence.Content {
				positive = false
				break
			}
		}
	}

	type RetIntent struct {
		ID       int64  `json:"id"`
		Name     string `json:"name"`
		Positive bool   `json:"positive"`
	}
	controllers.ReturnOK(w, &RetIntent{
		ID:       intent.ID,
		Name:     intent.Name,
		Positive: positive,
	})
}

func newRecordQuery(r *http.Request) (query *dataV2.VisitRecordsQuery, err error, errCode int) {
	request := dataV2.VisitRecordsRequest{}
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		errCode = data.ErrCodeInvalidRequestBody
		return
	}

	enterpriseID := requestheader.GetEnterpriseID(r)
	appID := requestheader.GetAppID(r)

	startTime := time.Unix(request.StartTime, 0).UTC()
	endTime := time.Unix(request.EndTime, 0).UTC()

	var platforms []string
	var genders []string

	if request.Platforms != nil {
		for _, platform := range request.Platforms {
			p, found := services.GetTagNameByID(appID, "platform", platform)
			if !found {
				err = data.ErrInvalidRequestBody
				errCode = data.ErrCodeInvalidParameterPlatform
				return
			}

			platforms = append(platforms, p)
		}
	}

	if request.Genders != nil {
		for _, gender := range request.Genders {
			g, found := services.GetTagNameByID(appID, "sex", gender)
			if !found {
				err = data.ErrInvalidRequestBody
				errCode = data.ErrCodeInvalidParameterGender
				return
			}

			genders = append(genders, g)
		}
	}

	var page, limit int64

	if request.Page == nil {
		page = dataCommon.RecordsDefaultPage
	} else {
		page = *request.Page
	}

	if request.Limit == nil {
		limit = dataCommon.RecordsDefaultPageLimit
	} else {
		limit = *request.Limit
	}

	query = &dataV2.VisitRecordsQuery{
		CommonQuery: data.CommonQuery{
			EnterpriseID: enterpriseID,
			AppID:        appID,
			StartTime:    startTime,
			EndTime:      endTime,
		},
		Modules:       request.Modules,
		Platforms:     platforms,
		Genders:       genders,
		Emotions:      request.Emotions,
		IsIgnored:     request.IsIgnored,
		IsMarked:      request.IsMarked,
		Keyword:       request.Keyword,
		UserID:        request.UserID,
		SessionID:     request.SessionID,
		TESessionID:   request.TESessionID,
		Intent:        request.Intent,
		MinScore:      request.MinScore,
		MaxScore:      request.MaxScore,
		LowConfidence: request.LowConfidence,
		FaqCategories: request.FaqCategories,
		FaqRobotTags:  request.FaqRobotTags,
		Feedback:      request.Feedback,
		From:          (page - 1) * int64(limit),
		Limit:         limit,
	}

	if query.From+query.Limit > data.ESMaxResultWindow {
		err = data.ErrInvalidRequestBody
		errCode = data.ErrCodeInvalidParameterPage
		return
	}

	return
}

func RecordsIntentMarkHandler(w http.ResponseWriter, r *http.Request) {
	type internalError struct {
		IsRollbacked bool `json:"rollbacked"`
	}
	var request struct {
		IntentID int64    `json:"intent"`
		Positive *bool    `json:"positive"`
		Records  []string `json:"records"`
	}
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		controllers.ReturnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, err.Error()))
		return
	}
	if request.IntentID != -1 && request.Positive == nil {
		controllers.ReturnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, "positive is required."))
		return
	}
	if request.Records == nil {
		controllers.ReturnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, "records is required"))
		return
	}

	appid := requestheader.GetAppID(r)
	q := &dataV2.VisitRecordsQuery{
		CommonQuery: data.CommonQuery{
			AppID: appid,
		},
		RecordIDs: make([]interface{}, len(request.Records)),
		Limit:     int64(len(request.Records)),
	}

	for i, r := range request.Records {
		q.RecordIDs[i] = r
	}

	// Need retrive record's userq
	result, err := servicesV2.VisitRecordsQuery(q)
	if err != nil {
		logger.Error.Printf("check record content failed, %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(internalError{
			IsRollbacked: true,
		})
		return
	}

	sentenceType := intentengine.TypePositive
	if request.Positive != nil && !(*request.Positive) {
		sentenceType = intentengine.TypeNegative
	}
	existed := map[string]bool{}
	updateSentences := []string{}
	for _, record := range result.Data {
		// ignore duplicated sentence
		if _, ok := existed[record.UserQ]; ok {
			continue
		}

		existed[record.UserQ] = true
		updateSentences = append(updateSentences, record.UserQ)
	}

	if request.IntentID > 0 {
		err = intentengine.AddIntentSentences(appid, request.IntentID, sentenceType, updateSentences)
	} else {
		err = intentengine.DelIntentSentences(appid, request.IntentID, updateSentences)
	}
	if err != nil {
		logger.Error.Printf("update intent failed, %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(internalError{
			IsRollbacked: true,
		})
		return
	}

	newQuery := &dataV2.VisitRecordsQuery{
		CommonQuery: data.CommonQuery{
			AppID: appid,
		},
		RecordIDs: make([]interface{}, len(request.Records)),
	}

	for i, r := range request.Records {
		newQuery.RecordIDs[i] = r
	}

	err = servicesV2.UpdateRecords(newQuery, servicesCommon.UpdateRecordIntentMark(request.IntentID))
	if err != nil {
		logger.Error.Printf("service update record failed, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(internalError{
			IsRollbacked: true,
		})
		return
	}

	type response struct {
		Done []string `json:"done"`
		Skip []string `json:"skip"`
	}
	w.WriteHeader(http.StatusOK)
	resp := response{
		Done: request.Records,
		Skip: []string{},
	}
	data, _ := json.Marshal(resp)
	w.Write(data)
}
