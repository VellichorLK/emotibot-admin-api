package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime/pprof"
	"strings"

	"emotibot.com/emotigo/module/vipshop-admin/FAQ"
	"emotibot.com/emotigo/module/vipshop-admin/util"
	_ "github.com/go-sql-driver/mysql"
)

var APPID, remoteHost string
var volumePath = "."
var dbName = "emotibot"
var user = "root"
var pass = "password"

func init() {
	var found bool
	APPID, found = os.LookupEnv("APPID")
	if !found {
		log.Fatalln("ENV APPID not found")
	}
	remoteHost, found = os.LookupEnv("host")
	if !found {
		log.Fatalln("ENV host not found")
	}
	v, found := os.LookupEnv("volume")
	if !found {
		fmt.Println("ENV volume not found, use default volume: " + volumePath)
	} else {
		fmt.Println("volume set for " + v)
		volumePath = v
	}

}
func main() {
	linkURL := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true",
		user,
		pass,
		remoteHost,
		dbName,
	)
	var err error
	db, err := sql.Open("mysql", linkURL)
	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		panic(err)
	}
	util.SetDB("main", db)
	defer db.Close()

	var successSQ = make(map[string]int)
	var failedSQ = make(map[string][]int)
	var failedStdQ = make(map[string]struct{})
	categories, err := FAQ.NewCategoryTree(APPID)
	if err != nil {
		panic(err)
	}
	f, err := os.Create(volumePath + "/pprof")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	var size int
	err = db.QueryRow("SELECT count(SQ_Id) FROM " + APPID + "_squestion WHERE status != 1").Scan(&size)
	if err != nil {
		panic(err)
	}
	if size > 0 {
		fmt.Println("SQ DB is under modify, stop script.")
		return
	}
	err = db.QueryRow("SELECT count(*) FROM " + APPID + "_squestion").Scan(&size)
	if err != nil {
		panic(err)
	}
	rows, err := db.Query("SELECT SQ_Id, Question_Id,Content FROM " + APPID + "_squestion ")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var count = 0
	for rows.Next() {
		var sqId, qId int
		var content string
		count++

		fmt.Printf("\rOn %d/%d", count, size)

		rows.Scan(&sqId, &qId, &content)
		query := fmt.Sprintf("id:vipshop_other_%d_%d", qId, sqId)
		resp, err := SolrQuery(remoteHost, query)
		if err != nil {
			panic(err)
		}
		if resp.Num != 1 {
			if _, ok := successSQ[content]; ok {
				continue
			}
			if stdQs, ok := failedSQ[content]; ok {
				failedSQ[content] = append(stdQs, qId)
			} else {
				failedSQ[content] = []int{qId}
			}
		} else {
			if resp.Docs[0].OriginalSentence != content {
				fmt.Printf("Fatal Error: Solr doc is not match with content [%s:%s], Hit enter to continue\n", resp.Docs[0].OriginalSentence, content)
				reader := bufio.NewReader(os.Stdin)
				reader.ReadString('\n')
				continue
			}
			successSQ[content] = qId
			delete(failedSQ, content)
		}
	}
	fmt.Println("Finish SQ Part")
	err = db.QueryRow("SELECT count(Question_Id) FROM " + APPID + "_question WHERE Status != 0").Scan(&size)
	if err != nil {
		panic(err)
	}
	rows, err = db.Query("SELECT Question_Id, Content, CategoryId FROM " + APPID + "_question")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var qId int
		var content string
		err = rows.Scan(&qId, &content)
		if err != nil {
			panic(err)
		}
		query := fmt.Sprintf("id:vipshop_other_%d_%d_self", qId, qId)
		resp, err := SolrQuery(remoteHost, query)
		if err != nil {
			panic(err)
		}
		if resp.Num != 1 {
			failedStdQ[content] = struct{}{}
		} else {
			delete(failedSQ, content)
		}
	}

	fmt.Printf("Num of %d sq scanned\n", count)
	fmt.Printf("found %d num of success cases\n", len(successSQ))
	fmt.Printf("found %d num of bad sq\n", len(failedSQ))
	err = ioutil.WriteFile("./badcase", []byte(fmt.Sprintf("BadSQ:%+v\nBadStdQ:%+v", failedSQ, failedStdQ)), 0644)
	if err != nil {
		fmt.Println(err)
	}

	var input string
	fmt.Println("Do you want to repair failed data?(Y/N)")
	reader := bufio.NewReader(os.Stdin)
	input, _ = reader.ReadString('\n')
	if strings.ToLower(input) == "y" {

	}
}

type solrResponse struct {
	Num  int       `json:"numFound"`
	Docs []solrDoc `json:"docs"`
}

type solrDoc struct {
	ID               string `json:"id"`
	OriginalSentence string `json:"sentence_original"`
}

//SolrQuery use for solr query api
func SolrQuery(host string, query string) (solrResponse, error) {
	req, err := http.NewRequest(http.MethodGet, "http://"+host+":8081/solr/3rd_core/select", nil)
	if err != nil {
		panic(err)
	}
	val := make(url.Values, 2)
	val.Set("q", query)
	val.Set("wt", "json")
	req.URL.RawQuery = val.Encode()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	var sResp struct {
		Response solrResponse `json:"response"`
	}
	err = json.Unmarshal(data, &sResp)
	if err != nil {
		fmt.Println(string(data))
		panic(err)
	}

	return sResp.Response, nil
}
