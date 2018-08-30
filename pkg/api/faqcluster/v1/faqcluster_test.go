package faqcluster

import (
	"context"
	"database/sql"
	"io/ioutil"
	"net/url"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func TestNewClient(t *testing.T) {
	type testCase struct {
		Address           string
		ExpectClusterAddr string
		ExpectResultAddr  string
	}
	testcases := map[string]testCase{
		"normal": testCase{
			Address:           "http://127.0.0.1",
			ExpectClusterAddr: "http://127.0.0.1/clustering/",
		},
		"custom port": testCase{
			Address:           "http://172.17.0.1:13014",
			ExpectClusterAddr: "http://172.17.0.1:13014/clustering/",
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(tt *testing.T) {
			addr, _ := url.Parse(tc.Address)
			client := NewClient(*addr)
			if client.clusterEndpoint != tc.ExpectClusterAddr {
				tt.Fatalf("expect cluster endpoint to be %s but got %s", tc.ExpectClusterAddr, client.clusterEndpoint)
			}
		})
	}
}

func TestIntergratedAPI(t *testing.T) {
	db, _ := sql.Open("mysql", "root:password@tcp(172.16.101.98:3306)/backend_log?parseTime=true&loc=Asia%2FShanghai")
	rows, _ := db.Query("SELECT id, user_q FROM records LIMIT 30")
	defer rows.Close()
	var data = make([]interface{}, 0)
	for rows.Next() {
		var id, userQ string
		rows.Scan(&id, &userQ)
		data = append(data, map[string]string{
			"id":    id,
			"value": userQ,
		})
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	addr, _ := url.Parse("http://127.0.0.1:13014")
	var client = NewClient(*addr)
	ctx := context.Background()
	paramas := map[string]interface{}{
		"model_version": "unknown_20180830143445",
	}
	result, err := client.Clustering(ctx, paramas, data)
	if rawErr, ok := err.(*RawError); ok {
		errData, _ := ioutil.ReadAll(rawErr.Body)
		t.Fatalf("client error %s with request para %s, raw body %s", rawErr.Error(), rawErr.Input, errData)
	}
	if err != nil {
		t.Fatalf("do cluster failed, %v", err)
	}
	if len(result.Clusters) != 0 {
		t.Fatal("expect clustering result > 0 ")
	}
}
