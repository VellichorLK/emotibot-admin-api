package qi

import (
	"encoding/json"
	"reflect"
	"testing"
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
