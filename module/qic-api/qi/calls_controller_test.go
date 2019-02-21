package qi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"emotibot.com/emotigo/module/admin-api/util/requestheader"

	"github.com/gorilla/mux"

	"emotibot.com/emotigo/module/qic-api/model/v1"
	"github.com/stretchr/testify/assert"
)

func TestNewCallReq_UnmarshalJSON(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    NewCallReq
		wantErr bool
	}{
		{
			name: "custom columns",
			args: args{
				data: []byte(`{"file_name":"1.mp3","testing":"gg"}`),
			},
			want: NewCallReq{
				FileName: "1.mp3",
				CustomColumns: map[string]interface{}{
					"testing": "gg",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got NewCallReq
			err := json.Unmarshal(tt.args.data, &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCallReq.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got = %v, want = %v", got, tt.want)
			}
		})
	}
}

func Test_parseJSONKeys(t *testing.T) {
	type args struct {
		n interface{}
	}
	tests := []struct {
		name string
		args args
		want map[string]struct{}
	}{
		{
			name: "simple json",
			args: args{
				n: struct {
					A string `json:"a"`
					B int    `json:"b"`
				}{},
			},
			want: map[string]struct{}{
				"a": {},
				"b": {},
			},
		},
		{
			name: "json with omitempty",
			args: args{
				n: struct {
					A string `json:"a,omitempty"`
				}{},
			},
			want: map[string]struct{}{
				"a": {},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseJSONKeys(tt.args.n); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseJSONKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCallResp_MarshalJSON(t *testing.T) {

	tests := []struct {
		name    string
		arg     CallResp
		want    []byte
		wantErr bool
	}{
		{
			name: "without custom columns",
			arg: CallResp{
				CallID: 1,
				Status: 2,
			},
			want: []byte(`{"call_id":1,"call_time":0,"deal":0,"status":2,"upload_time":0,"duration":0,"left_silence_time":0,"right_silence_time":0,"left_speed":null,"right_speed":null}`),
		},
		{
			name: "with custom columns",
			arg: CallResp{
				CallID: 1,
				CustomColumns: map[string]interface{}{
					"location": "taipei",
				},
			},
			want: []byte(`{"call_id":1,"call_time":0,"deal":0,"status":0,"upload_time":0,"duration":0,"left_silence_time":0,"right_silence_time":0,"left_speed":null,"right_speed":null, "location": "taipei"}`),
		},
		{
			name: "with overlapped columns",
			arg: CallResp{
				CallID: 1,
				CustomColumns: map[string]interface{}{
					"call_id": 10,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("CallResp.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			var (
				g = map[string]interface{}{}
				w = map[string]interface{}{}
			)
			json.Unmarshal(got, &g)
			json.Unmarshal(tt.want, &w)
			assert.Equal(t, w, g, "")
		})
	}
}

func Test_callRequest(t *testing.T) {
	//dumpNext always write call id back to body.
	dumpNext := func(w http.ResponseWriter, r *http.Request, c *model.Call) {
		fmt.Fprintf(w, "[%d]", c.ID)
	}
	tmp := call
	defer func() {
		call = tmp
	}()
	// mock call always return an empty call with given callID.
	call = func(callID int64, enterprise string) (c model.Call, err error) {
		return model.Call{ID: callID}, nil
	}
	type args struct {
		callID string
		header map[string]string
	}
	tests := []struct {
		name       string
		args       args
		wantBody   []byte
		wantStatus int
	}{
		{
			name: "success condition",
			args: args{
				callID: "1",
				header: map[string]string{
					requestheader.ConstEnterpriseIDHeaderKey: "csbot",
				},
			},
			wantBody:   []byte(`[1]`),
			wantStatus: 200,
		},
		{
			name: "invalid call_id",
			args: args{
				callID: "notanumber",
			},
			wantBody:   []byte("{\"status\":-3,\"message\":\"请求参数错误: invalid call_id 'notanumber', need to be int\",\"result\":null}"),
			wantStatus: 400,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := callRequest(dumpNext)
			sm := mux.NewRouter()
			sm.HandleFunc("/test/{call_id}", h).Methods("GET").Name("testing")
			w := httptest.NewRecorder()
			addr, _ := sm.Get("testing").URL("call_id", tt.args.callID)
			r := httptest.NewRequest("GET", addr.String(), nil)
			for n, v := range tt.args.header {
				r.Header.Set(n, v)
			}
			sm.ServeHTTP(w, r)
			t.Log(tt)
			assert.Equal(t, tt.wantStatus, w.Code, "status code must be equal")
			body, _ := ioutil.ReadAll(w.Body)
			assert.JSONEq(t, string(tt.wantBody), string(body))
		})
	}
}
