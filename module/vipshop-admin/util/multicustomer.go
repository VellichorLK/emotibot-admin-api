package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"

	"emotibot.com/emotigo/module/vipshop-admin/ApiError"
)

const (
	MulticustomerURLKey = "MC_URL"
)

//DefaultMCClient is used as default client for MC, like http.DefaultClient.
var DefaultMCClient MultiCustomerClient = MultiCustomerHttpClient{}

type MultiCustomerClient interface {
	McImportExcel(fileHeader multipart.FileHeader, UserID string, UserIP string, mode string, appid string) (MCResponse, error)
	McExportExcel(UserID string, UserIP string, AnswerIDs []string, appid string) (MCResponse, error)
	McManualBusiness(appid string) (int, error)
}

type MultiCustomerHttpClient http.Client

func UpdateWordBank(appid string, userID string, userIP string, retFile string) (int, error) {
	mcURL := getGlobalEnv(MulticustomerURLKey)
	// http://172.16.101.47:14501/entity?
	// app_id, userip, userid, file_name
	reqURL := fmt.Sprintf("%s/entity?app_id=%s&userid=%s&userip=%s&file_name=%s", mcURL, appid, userID, userIP, retFile)
	logTraceMC("req", reqURL)

	body, resErr := HTTPGetSimpleWithTimeout(reqURL, 5)
	if resErr != nil {
		logMCError(resErr)
		return ApiError.DICT_SERVICE_ERROR, resErr
	}
	logTraceMC("update wordbank", body)
	return ApiError.SUCCESS, nil
}

func McUpdateFunction(appid string) (int, error) {
	mcURL := getGlobalEnv(MulticustomerURLKey)
	// robot_config?
	// app_id, type=function
	// $curl = "curl '" . SETTING_API . $appid . "&type=function' >> /dev/null &";
	reqURL := fmt.Sprintf("%s/robot_config?app_id=%s&type=function", mcURL, appid)
	logTraceMC("req", reqURL)

	body, resErr := HTTPGetSimpleWithTimeout(reqURL, 30)
	if resErr != nil {
		logMCError(resErr)
		return ApiError.DICT_SERVICE_ERROR, resErr
	}
	logTraceMC("update function", body)
	return ApiError.SUCCESS, nil
}

func McRebuildRobotQA(appid string) (int, error) {
	mcURL := getGlobalEnv(MulticustomerURLKey)
	// manual_edit
	// app_id
	reqURL := fmt.Sprintf("%s/manual_edit?app_id=%s&type=robot", mcURL, appid)
	logTraceMC("req", reqURL)

	body, resErr := HTTPGetSimpleWithTimeout(reqURL, 30)
	if resErr != nil {
		logMCError(resErr)
		return ApiError.DICT_SERVICE_ERROR, resErr
	}
	logTraceMC("rebuild robotQA", body)
	return ApiError.SUCCESS, nil
}

//McManualBusiness use DefaultMCClient to call REST API of Multicustomer , response for scanning faq into Solr
func McManualBusiness(appid string) (int, error) {
	return DefaultMCClient.McManualBusiness(appid)
}

//McManualBusiness is a REST API of Multicustomer, response for scanning faq into Solr
func (m MultiCustomerHttpClient) McManualBusiness(appid string) (int, error) {
	mcURL := getGlobalEnv(MulticustomerURLKey)
	// manual_edit
	// app_id
	reqURL := fmt.Sprintf("%s/manual_business?app_id=%s&type=other", mcURL, appid)
	logTraceMC("req", reqURL)

	response, err := http.Get(reqURL)
	if err != nil {
		logMCError(err)
		return ApiError.DICT_SERVICE_ERROR, err
	}
	defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logMCError(err)
		return ApiError.IO_ERROR, err
	}
	logTraceMC("rebuild question", string(data))
	return ApiError.SUCCESS, nil
}

func logTraceMC(function string, msg string) {
	LogTrace.Printf("[MC][%s]:%s", function, msg)
}

func logMCError(err error) {
	logTraceMC("connect error", err.Error())
}

// McImportExcel is a REST API of Multicustomer module, response for importing xlsx
func (m MultiCustomerHttpClient) McImportExcel(fileHeader multipart.FileHeader, userID string, userIP string, mode string, appid string) (MCResponse, error) {
	mcURL := getGlobalEnv(MulticustomerURLKey)
	w := &bytes.Buffer{}
	writer := multipart.NewWriter(w)
	fw, err := writer.CreateFormFile("file", fileHeader.Filename)
	f, err := fileHeader.Open()
	defer f.Close()
	// data, _ := ioutil.ReadAll(f)
	if _, err = io.Copy(fw, f); err != nil {
		return MCResponse{}, err
	}
	writer.Close()
	queryString := url.Values{}
	queryString.Set("app_id", appid)
	queryString.Set("type", "other")
	queryString.Set("userid", userID)
	queryString.Set("userip", userIP)
	queryString.Set("module", mode)

	reqURL := fmt.Sprintf("%s/business?%s", mcURL, queryString.Encode())
	req, err := http.NewRequest("POST", reqURL, w)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	res, err := http.DefaultClient.Do(req)

	var response MCResponse

	if err != nil {
		return response, err
	}

	switch res.StatusCode {
	case http.StatusOK:
		data, _ := ioutil.ReadAll(res.Body)
		fmt.Println(string(data))
		err = json.Unmarshal(data, &response)
		return response, err
	case http.StatusBadRequest:
		body, _ := ioutil.ReadAll(res.Body)
		LogInfo.Println(string(body))
		return response, errors.New("Multicustomer return Bad Request")
	case http.StatusServiceUnavailable:
		data, _ := ioutil.ReadAll(res.Body)
		err = json.Unmarshal(data, &response)
		if err != nil {
			return response, err
		}
		return response, ErrorMCLock
	default:
		return response, errors.New("Multicustomer return " + string(res.StatusCode) + ", no operation will do")
	}

}

// MCExportExcel
func (m MultiCustomerHttpClient) McExportExcel(userID string, userIP string, answerIDs []string, appid string) (MCResponse, error) {
	var mcResponse MCResponse
	mcURL := getGlobalEnv(MulticustomerURLKey)
	type requestJSON struct {
		AppID     string   `json:"app_id"`
		UserID    string   `json:"userid"`
		UserIP    string   `json:"userip"`
		Module    string   `json:"module"`
		AnswerIDs []string `json:"answerid"`
	}
	reqBody := requestJSON{}
	reqBody.AppID = appid
	reqBody.UserID = userID
	reqBody.UserIP = userIP
	reqBody.Module = "business"
	reqBody.AnswerIDs = answerIDs

	bodyStr, err := json.Marshal(reqBody)
	if err != nil {
		return mcResponse, err
	}

	reqURL := fmt.Sprintf("%s/download", mcURL)
	req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewBuffer(bodyStr))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		return mcResponse, err
	}
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return mcResponse, err
	}
	switch response.StatusCode {
	case http.StatusOK:
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return mcResponse, err
		}
		err = json.Unmarshal(body, &mcResponse)
		return mcResponse, err
	case http.StatusServiceUnavailable:
		return mcResponse, ErrorMCLock
	default:
		return mcResponse, errors.New("Multicustomer return " + string(response.StatusCode) + ", no operation will do")
	}
}

func checkMCDB() {

}

//MCResponse represent partical json struct returned by MultiCustomer
type MCResponse struct {
	SyncInfo struct {
		StatID int    `json:"stateID"`
		UserID string `json:"userID"`
		Action string `json:"action"`
	} `json:"sync_info"`
}
