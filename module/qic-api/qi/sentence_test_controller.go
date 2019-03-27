package qi

import (
	"net/http"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/pkg/logger"
	"strconv"
)

type TestSentenceReq struct {
	Name       string `json:"name"`
	CategoryID uint64 `json:"category_id,string"`
	TestedID   uint64 `json:"tested_id"`
}

type TestResult struct {
	Index     int      `json:"index"`
	Name      string   `json:"name"`
	HitTags   []uint64 `json:"hit_tags"`
	FailTags  []uint64 `json:"fail_tags"`
	MatchText []string `json:"match_text"`
	Accuracy  float32  `json:"accuracy"`
}

type testSenResp struct {
	UUID string `json:"uuid"`
}

func handlePredictSentences(w http.ResponseWriter, r *http.Request) {
	enterpriseID := requestheader.GetEnterpriseID(r)

	err := PredictSentences(enterpriseID)
	if err != nil {
		logger.Error.Printf("fail to predict sentence. %s \n", err.Error())
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.IO_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
	util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.SUCCESS, nil), http.StatusOK)
}

func handleGetTestSentences(w http.ResponseWriter, r *http.Request) {
	enterpriseID := requestheader.GetEnterpriseID(r)

	idStr := parseID(r)
	testedID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "need valid testedID"), http.StatusBadRequest)
		return
	}

	testSentences, err := GetTestSentences(enterpriseID, testedID)
	if err != nil {
		logger.Error.Printf("fail to get sentences. %s \n", err.Error())
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
	if err = util.WriteJSON(w, testSentences); err != nil {
		logger.Error.Println(err.Error())
	}
}

func handleNewTestSentence(w http.ResponseWriter, r *http.Request) {
	enterpriseID := requestheader.GetEnterpriseID(r)

	var testSenReq TestSentenceReq
	err := util.ReadJSON(r, &testSenReq)
	if err != nil {
		logger.Error.Printf("fail to parse request. %s \n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	uuidStr, err := NewTestSentence(enterpriseID, &testSenReq)
	if err != nil {
		logger.Error.Printf("fail to insert testSentence. %s \n", err.Error())
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	if err = util.WriteJSON(w, testSenResp{UUID: *uuidStr}); err != nil {
		logger.Error.Println(err.Error())
	}
}

func handleDeleteTestSentence(w http.ResponseWriter, r *http.Request) {
	enterpriseID := requestheader.GetEnterpriseID(r)
	uuidStr := parseID(r)

	affected, err := SoftDeleteTestSentence(enterpriseID, uuidStr)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
	if affected == 0 {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "No record is deleted"), http.StatusBadRequest)
		return
	}
}

func handleGetSentenceTestResult(w http.ResponseWriter, r *http.Request) {
	enterpriseID := requestheader.GetEnterpriseID(r)
	categoryID, err := getQueryCategoryID(r)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err), http.StatusBadRequest)
		return
	}

	testResults, err := GetSentenceTestResult(enterpriseID, categoryID)
	if err != nil {
		logger.Error.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = util.WriteJSON(w, testResults); err != nil {
		logger.Error.Println(err.Error())
	}
}

func handleGetSentenceTestOverview(w http.ResponseWriter, r *http.Request) {
	enterpriseID := requestheader.GetEnterpriseID(r)

	senTestOverview, err := GetSentenceTestOverview(enterpriseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = util.WriteJSON(w, senTestOverview); err != nil {
		logger.Error.Println(err.Error())
	}
}

func handleGetSentenceTestDetail(w http.ResponseWriter, r *http.Request) {
	enterpriseID := requestheader.GetEnterpriseID(r)
	idStr := parseID(r)

	senID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	detail, err := GetSentenceTestDetail(enterpriseID, senID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = util.WriteJSON(w, detail); err != nil {
		logger.Error.Println(err.Error())
	}
}
