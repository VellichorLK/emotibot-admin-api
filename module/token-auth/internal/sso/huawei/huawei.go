package huawei

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"emotibot.com/emotigo/module/token-auth/internal/util"
)

// HuaweiSSO is used for sso system in huawei
type HuaweiSSO struct {
	Config *util.SSOConfig
}

type token struct {
	Login  string `json:"hwsso_login"`
	T      string `json:"hwssot"`
	TINTER string `json:"hwssotiner3"`
	UID    string `json:"login_uid"`
}

type validateInput struct {
	Token *token `json:"token"`
	URL   string `json:"url"`
}

type ssoUser struct {
	UID string `json:"uid"`
}

type validateReturn struct {
	Code int      `json:"errorCode"`
	Msg  string   `json:"errorMsg"`
	User *ssoUser `json:"user"`
}

func (handler *HuaweiSSO) LoadConfig(config *util.SSOConfig) error {
	if config == nil {
		return errors.New("Invalid sso config")
	}
	handler.Config = &util.SSOConfig{
		ValidateURL: config.ValidateURL,
		LoginURL:    config.LoginURL,
		LogoutURL:   config.LogoutURL,
	}
	util.LogTrace.Printf("Load huawei sso with config: %+v\n", handler.Config)
	return nil
}

func (handler *HuaweiSSO) ValidateRequest(r *http.Request) (string, string, error) {
	requestInfo := validateInput{
		Token: &token{},
		URL:   r.Referer(),
	}
	requestInfo.Token = &token{}

	cookies := r.Cookies()
	for _, cookie := range cookies {
		switch cookie.Name {
		case "hwsso_login":
			requestInfo.Token.Login = cookie.Value
		case "hwssot":
			requestInfo.Token.T = cookie.Value
		case "hwssotiner3":
			requestInfo.Token.TINTER = cookie.Value
		case "login_uid":
			requestInfo.Token.UID = cookie.Value
		}
	}

	data, _ := json.Marshal(requestInfo)
	util.LogTrace.Printf("SSO Validate input: %s\n", data)
	ssoUser, err := handler.CallValidate(&requestInfo)
	if err != nil {
		return "", "", err
	}

	return ssoUser.UID, "user_name", nil
}

func (handler *HuaweiSSO) CallValidate(requestInfo *validateInput) (*ssoUser, error) {
	data, err := json.Marshal(requestInfo)
	if err != nil {
		util.LogError.Printf("json.Marshal() failed with '%s'\n", err)
		return nil, err
	}
	client := &http.Client{}
	client.Timeout = time.Second * 15

	uri := handler.Config.ValidateURL
	body := bytes.NewBuffer(data)
	req, err := http.NewRequest(http.MethodPut, uri, body)
	if err != nil {
		util.LogError.Printf("http.NewRequest() failed with '%s'\n", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := client.Do(req)
	if err != nil {
		util.LogError.Printf("client.Do() failed with '%s'\n", err)
		return nil, err
	}

	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		util.LogError.Printf("ioutil.ReadAll() failed with '%s'\n", err)
		return nil, err
	}
	util.LogTrace.Printf("Receive from SSO validate: %s\n", content)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SSO Status is %d", resp.StatusCode)
	}

	ssoUser := ssoUser{}
	ret := validateReturn{
		User: &ssoUser,
	}
	err = json.Unmarshal(content, &ret)
	if err != nil {
		util.LogError.Println("Unmarshal sso user fail,", err.Error())
	}
	return &ssoUser, nil
}

func (handler *HuaweiSSO) ValidateDebug(r *http.Request) string {
	ret := ""
	requestInfo := validateInput{
		Token: &token{},
		URL:   r.Referer(),
	}
	requestInfo.Token = &token{}

	cookies := r.Cookies()
	for _, cookie := range cookies {
		switch cookie.Name {
		case "hwsso_login":
			requestInfo.Token.Login = cookie.Value
		case "hwssot":
			requestInfo.Token.T = cookie.Value
		case "hwssotiner3":
			requestInfo.Token.TINTER = cookie.Value
		case "login_uid":
			requestInfo.Token.UID = cookie.Value
		}
	}

	data, _ := json.Marshal(requestInfo)
	ret = fmt.Sprintf("SSO Validate input: %s\n", data)
	ssoUser, err := handler.CallValidate(&requestInfo)
	if err != nil {
		ret = ret + fmt.Sprintf("SSO validate fail: %s\n", err.Error())
	} else {
		ret = ret + fmt.Sprintf("Get SSO user, get user which user_name is %s\n", ssoUser.UID)
	}
	return ret
}
