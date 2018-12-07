package feedback

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"
	"github.com/gorilla/mux"
)

func TestGetReasonHandler(t *testing.T) {
	getReasonService = GenMockServiceGetReasons(t, "csbot")
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-AppID", "csbot")

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleGetFeedbackReasons)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestAddReasonHandler(t *testing.T) {
	addReasonService = GenMockServiceAddReason(t, "csbot", "reason1")
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("POST", "/", bytes.NewBufferString("{\"content\": \"reason1\"}"))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-AppID", "csbot")

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleAddFeedbackReason)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestDeleteReasonHandler(t *testing.T) {
	delReasonService = GenMockServiceDeleteReason(t, "csbot", 10)

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	router := mux.NewRouter()
	router.HandleFunc("/{id}", handleDeleteFeedbackReason)
	ts := httptest.NewServer(router)
	defer ts.Close()

	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/10", ts.URL), nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-AppID", "csbot")

	c := &http.Client{}
	_, e := c.Do(req)
	if e != nil {
		t.Error(e)
	}
}

func GenMockServiceGetReasons(t *testing.T, except string) func(appid string) ([]*Reason, AdminErrors.AdminError) {
	return func(appid string) ([]*Reason, AdminErrors.AdminError) {
		if appid != except {
			t.Errorf("Excepted %s, but appid is %s", except, appid)
		}
		return nil, nil
	}
}
func GenMockServiceAddReason(t *testing.T, exceptAppID string, exceptContent string) func(appid string, content string) (*Reason, AdminErrors.AdminError) {
	return func(appid string, content string) (*Reason, AdminErrors.AdminError) {
		if appid != exceptAppID {
			t.Errorf("Excepted %s, but appid is %s", exceptAppID, appid)
		}
		if content != exceptContent {
			t.Errorf("Excepted %s, but content is %s", exceptContent, content)
		}
		return nil, nil
	}
}
func GenMockServiceDeleteReason(t *testing.T, exceptAppID string, exceptID int64) func(appid string, id int64) AdminErrors.AdminError {
	return func(appid string, id int64) AdminErrors.AdminError {
		if appid != exceptAppID {
			t.Errorf("Excepted %s, but appid is %s", exceptAppID, appid)
		}
		if id != exceptID {
			t.Errorf("Excepted %d, but content is %d", exceptID, id)
		}
		return nil
	}
}
