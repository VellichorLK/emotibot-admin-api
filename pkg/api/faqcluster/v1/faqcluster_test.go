package faqcluster

import (
	"context"
	"encoding/csv"
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
	type testCase struct {
		input              string
		expectClustersSize int
		expectRemovedSize  int
		expectTotalSize    int
	}
	f, err := os.Open("./testdata/outputexample.csv")
	if err != nil {
		t.Fatal(err)
	}
	reader := csv.NewReader(f)
	rows, _ := reader.ReadAll()
	var testcases = map[string]testCase{}
	for _, row := range rows {
		var t testCase
		name := row[0]
		t.input = row[1]
		t.expectClustersSize, _ = strconv.Atoi(row[2])
		t.expectRemovedSize, _ = strconv.Atoi(row[3])
		t.expectTotalSize, _ = strconv.Atoi(row[4])
		testcases[name] = t
	}
	for name, tc := range testcases {
		t.Run(name, func(tt *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, tc.input)
			}))
			defer server.Close()
			addr, _ := url.Parse(server.URL)
			client := NewClientWithHTTPClient(addr, server.Client())
			result, err := client.Clustering(context.Background(), map[string]interface{}{}, nil)
			if err != nil {
				tt.Fatal("clustering err: ", err)
			}
			if len(result.Clusters) != tc.expectClustersSize {
				tt.Fatal("expect cluster size as ", tc.expectClustersSize, " but got ", len(result.Clusters))
			}
			if len(result.Filtered) != tc.expectRemovedSize {
				tt.Fatal("expect filtered size as ", tc.expectRemovedSize, " but got ", len(result.Filtered))
			}
			sum := 0
			for _, c := range result.Clusters {
				sum += len(c.Data)
			}
			sum += len(result.Filtered)
			if sum != tc.expectTotalSize {
				tt.Fatal("expect total sentences size to be ", tc.expectTotalSize, " but got ", sum)
			}
			server.Close()
		})
	}

}
func TestIntegratedAPI(t *testing.T) {
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
		if id >= 1000 {
			break
		}
	}

	addr, _ := url.Parse("http://172.16.101.98:13014")
	var client = NewClient(addr)
	ctx := context.Background()
	paramas := map[string]interface{}{
		"model_version": "unknown_20180830143445",
	}
	// logger.SetLevel("TRACE")
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
	sum := 0
	for _, c := range result.Clusters {
		sum += len(c.Data)
	}
	sum += len(result.Filtered)
	if sum != len(data) {
		t.Fatal("data sum expect to be ", len(data), " but got ", sum)
	}
	// t.Fatalf("%v", result.Clusters)
}
