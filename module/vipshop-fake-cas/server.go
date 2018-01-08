package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type ret_msg struct {
	Code int `json:"code"`
}

func main() {
	fmt.Println("Starting server...")
	http.HandleFunc("/auth", handler)
	fmt.Println(http.ListenAndServeTLS(":443", "./dist/server.crt", "./dist/server.key", nil))

}

var restrictedParameters = []string{"type", "appid", "ac", "pw"}

func handler(w http.ResponseWriter, r *http.Request) {
	parameters, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, "參數解讀錯誤,Error: %s \n", err.Error())
		return
	}
	var cantFound bool
	for _, key := range restrictedParameters {
		if _, found := parameters[key]; !found {
			cantFound = true
			break
		}
	}
	if len(parameters) != 4 || cantFound {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintln(w, "參數錯誤")
		return
	}

	if strings.Compare(parameters["appid"][0], "vca") != 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "appid incorrect")
		return
	}

	if strings.Compare(parameters["type"][0], "json") != 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "mock api only support json return type")
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	ret := ret_msg{
		Code: 200,
	}
	var acct = parameters["ac"][0]
	var password = parameters["pw"][0]
	log.Printf("login attempt from ip:%s ac:%s, pw:%s", r.RemoteAddr, acct, password)
	if !validateUser(acct, password) {
		log.Println("failed! ")
		ret.Code = 400
	} else {
		log.Println("success!")
	}
	retStr, _ := json.Marshal(ret)
	w.Write(retStr)
}

var valid_users = map[string]string{
	"user1":  "12345",
	"user2":  "12345",
	"user3":  "12345",
	"user4":  "12345",
	"user5":  "12345",
	"user6":  "12345",
	"user7":  "12345",
	"user8":  "12345",
	"user9":  "12345",
	"user10": "12345",
}

func validateUser(userID string, pw string) (isValid bool) {
	if password, ok := valid_users[userID]; ok && strings.Compare(password, pw) == 0 {
		isValid = true
	}

	return isValid
}
