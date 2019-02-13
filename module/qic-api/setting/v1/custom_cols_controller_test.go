package setting

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"emotibot.com/emotigo/module/admin-api/util/requestheader"

	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/general"
)

func restoreMockService() func() {
	tmp := defaultCustomColService
	return func() {
		defaultCustomColService = tmp
	}
}
func TestGetCustomColsHandler(t *testing.T) {
	defer restoreMockService()()
	type args struct {
		mock        func()
		QueryString url.Values
	}
	tests := []struct {
		name             string
		args             args
		expectStatusCode int
		expectBody       []byte
	}{
		{
			name: "200",
			args: args{
				mock: func() {
					defaultCustomColService.GetCustomCols = func(query model.UserKeyQuery) ([]CustomCol, general.Paging, error) {
						return []CustomCol{}, general.Paging{}, nil
					}
				},
				QueryString: url.Values{
					"limit": []string{"5"},
					"page":  []string{"1"},
				},
			},
			expectStatusCode: 200,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.mock()
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/test", nil)
			r.Header.Set(requestheader.ConstEnterpriseIDHeaderKey, "csbot")
			r.URL.RawQuery = tt.args.QueryString.Encode()
			GetCustomColsHandler(w, r)
			if w.Code != tt.expectStatusCode {
				body, _ := ioutil.ReadAll(w.Body)
				t.Log(string(body))
				t.Errorf("expect status code to be %d, but got %d", tt.expectStatusCode, w.Code)
			}
		})
	}
}
