package util

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// HTTPGetSimple is function to do HTTP GET request without param and timeout is
func HTTPGetSimple(url string) (string, error) {
	return HTTPGet(url, make(map[string]string), 0)
}

// HTTPGetSimpleWithTimeout is function to do HTTP GET request without param
func HTTPGetSimpleWithTimeout(url string, timeout int) (string, error) {
	return HTTPGet(url, make(map[string]string), timeout)
}

// HTTPGet is function to do HTTP GET request
func HTTPGet(url string, param map[string]string, timeout int) (string, error) {
	if url == "" {
		return "", errors.New("Invalid url")
	}

	var client *http.Client

	if timeout > 0 {
		getTimeout := time.Duration(time.Second) * time.Duration(timeout)
		client = &http.Client{
			Timeout: getTimeout,
		}
	} else {
		client = &http.Client{}
	}
	req, _ := http.NewRequest("GET", url, nil)

	query := req.URL.Query()
	for key, val := range param {
		query.Add(key, val)
	}
	req.URL.RawQuery = query.Encode()

	response, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func HTTPPost(url string, data interface{}, timeout int) (string, error) {
	if url == "" {
		return "", errors.New("Invalid url")
	}

	var client *http.Client

	if timeout > 0 {
		getTimeout := time.Duration(time.Second) * time.Duration(timeout)
		client = &http.Client{
			Timeout: getTimeout,
		}
	} else {
		client = &http.Client{}
	}

	jsonByte, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonByte)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	response, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func HTTPPut(url string, data interface{}, timeout int) (string, error) {
	if url == "" {
		return "", errors.New("Invalid url")
	}

	var client *http.Client

	if timeout > 0 {
		getTimeout := time.Duration(time.Second) * time.Duration(timeout)
		client = &http.Client{
			Timeout: getTimeout,
		}
	} else {
		client = &http.Client{}
	}

	jsonByte, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("PUT", url, strings.NewReader(string(jsonByte)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	response, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
