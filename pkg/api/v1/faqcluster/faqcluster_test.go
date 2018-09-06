package faqcluster

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"testing"
	"time"

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
			ExpectClusterAddr: "http://127.0.0.1/clustering/post",
			ExpectResultAddr:  "http://127.0.0.1/get_result",
		},
		"custom port": testCase{
			Address:           "http://172.17.0.1:13014",
			ExpectClusterAddr: "http://172.17.0.1:13014/clustering/post",
			ExpectResultAddr:  "http://172.17.0.1:13014/get_result",
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(tt *testing.T) {
			addr, _ := url.Parse(tc.Address)
			client := NewClient(*addr)
			if client.clusterEndpoint != tc.ExpectClusterAddr {
				tt.Fatalf("expect cluster endpoint to be %s but got %s", tc.ExpectClusterAddr, client.clusterEndpoint)
			}
			if client.resultEndpoint != tc.ExpectResultAddr {
				tt.Fatalf("expect result endpoint to be %s but got %s", tc.ExpectResultAddr, client.resultEndpoint)
			}
		})
	}
}

func TestIntergratedAPI(t *testing.T) {
	db, _ := sql.Open("mysql", "root:password@tcp(172.16.101.98:3306)/backend_log?parseTime=true&loc=Asia%2FShanghai")
	rows, _ := db.Query("SELECT id, user_q FROM records LIMIT 10000")
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
	resp, err := client.Clustering(data)
	if err != nil {
		log.Printf("response: %+v", resp)
		t.Fatal("got clustering error: ", err)
	}
	if resp.Status != StatusSuccess {
		log.Printf("response: %+v", resp)
		t.Fatal("expect response status to be ", StatusSuccess, "but got ", resp.Status)
	}
	var (
		result *Result
	)
	fmt.Println(resp.TaskID)
	for {
		result, err = client.GetResult(resp.TaskID)
		if err == nil {
			break
		} else if err != ErrNotDone {
			t.Fatalf("got result error, %v", err)
		}
		time.Sleep(time.Duration(2) * time.Second)
	}
	log.Printf("sucess result: %+v", result)
}
