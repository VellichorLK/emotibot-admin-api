package CAS

import (
	"crypto/tls"
	"errors"
	"io/ioutil"
	"net/http"
	"time"
)

//HTTPSGetRequest https request
// url:  https requst url
// timeout: second unit
// skipVerify: whether skip checking the certificate
func HTTPSGetRequest(url string, timeout int, skipVerify bool) (int, []byte, error) {
	if url == "" {
		return 0, nil, errors.New("Invalid url")
	}

	var getTimeout time.Duration
	if timeout > 0 {
		getTimeout = time.Duration(timeout) * time.Second
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipVerify},
	}

	client := &http.Client{Transport: transport, Timeout: getTimeout}

	response, err := client.Get(url)
	if err != nil {
		return 0, nil, err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, nil, err
	}

	return response.StatusCode, body, nil
}
