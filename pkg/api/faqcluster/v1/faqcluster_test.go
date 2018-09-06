package faqcluster

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"emotibot.com/emotigo/pkg/logger"
	_ "github.com/go-sql-driver/mysql"
)

var isIntegrationTest bool

func TestMain(m *testing.M) {
	flag.BoolVar(&isIntegrationTest, "it", false, "indicate integration test")
	flag.Parse()
	os.Exit(m.Run())
}
func TestNewClient(t *testing.T) {
	type testCase struct {
		Address           string
		ExpectClusterAddr string
		ExpectResultAddr  string
	}
	testcases := map[string]testCase{
		"normal": testCase{
			Address:           "http://127.0.0.1",
			ExpectClusterAddr: "http://127.0.0.1/clustering",
		},
		"custom port": testCase{
			Address:           "http://172.17.0.1:13014",
			ExpectClusterAddr: "http://172.17.0.1:13014/clustering",
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(tt *testing.T) {
			addr, _ := url.Parse(tc.Address)
			client := NewClient(addr)
			if client.clusterEndpoint != tc.ExpectClusterAddr {
				tt.Fatalf("expect cluster endpoint to be %s but got %s", tc.ExpectClusterAddr, client.clusterEndpoint)
			}
		})
	}
}

func TestOutputParsing(t *testing.T) {
	example := `{
		"errno":"success",
		"error_message":"",
		"para":{
			"model_version":"unknown_20180830143445",
			"deduplicate":false
		},
		"result":{
			"data":[
				{
					"centerQuestion":["1.6高，体重120穿多大号"],
					"clusterTag": ["測試"],
					"cluster":[{"id":"9","value":"1.6高，体重120穿多大号"}]
				},
				{
					"centerQuestion":["測試B"],
					"clusterTag": ["測試"],
					"cluster":[{"id":"1", "value":"測試B"}]
				}
			],
			"removed": [{"id":"1","value":"1"},{"id":"2","value":"1,76"},{"id":"12","value":"1.8床"}]
		}
	}
		`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, example)
	}))
	defer server.Close()
	addr, _ := url.Parse(server.URL)
	client := NewClientWithHTTPClient(addr, server.Client())
	result, err := client.Clustering(context.Background(), map[string]interface{}{}, nil)
	if err != nil {
		t.Fatal("err: ", err)
	}
	if len(result.Clusters) != 2 {
		t.Fatal("expect cluster size as 2 but got ", len(result.Clusters))
	}
	if tag := result.Clusters[0].Tags[0]; tag != "測試" {
		t.Fatal("expect tags to be 測試 but got ", tag)
	}
	if len(result.Filtered) != 3 {
		t.Fatal("expect filtered size as 3 but got ", len(result.Filtered))
	}

}
func TestIntergratedAPI(t *testing.T) {
	if !isIntegrationTest {
		t.Skip("integration test flag is not setted. skip it")
	}
	rawData, err := ioutil.ReadFile("./testdata/user_q.txt")
	if err != nil {
		t.Fatal(err)
	}
	var data = make([]interface{}, 0)
	rows := strings.Split(string(rawData), "\n")
	for id, row := range rows {
		data = append(data, map[string]string{
			"id":    strconv.Itoa(id),
			"value": row,
		})
		if id > 1000 {
			break
		}
	}

	addr, _ := url.Parse("http://172.16.101.98:13014")
	var client = NewClient(addr)
	ctx := context.Background()
	paramas := map[string]interface{}{
		"model_version": "unknown_20180830143445",
	}
	logger.SetLevel("TRACE")
	result, err := client.Clustering(ctx, paramas, data)
	if rawErr, ok := err.(*RawError); ok {
		t.Fatalf("client error %s with request para %s, raw body %s", rawErr.Error(), rawErr.Input, rawErr.Body)
	}
	if err != nil {
		t.Fatalf("do cluster failed, %v", err)
	}
	if len(result.Clusters) <= 0 {
		t.Fatal("expect clustering result > 0 ")
	}
	t.Fatalf("%v", result.Clusters)
}
