package main

import (
	"bufio"
	"bytes"
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

var APPID, remoteHost, solrHost string
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
	solrHost, found = os.LookupEnv("solr_host")
	if !found {
		solrHost = remoteHost
	}

	v, found := os.LookupEnv("volume")
	if !found {
		fmt.Println("ENV volume not found, use default volume: " + volumePath)
	} else {
		fmt.Println("volume set for " + v)
		volumePath = v
	}
	var devNull = &bytes.Buffer{}
	util.LogInit(devNull, devNull, devNull, devNull)
	err := util.LoadConfigFromFile("/app/.env")
	if err != nil {
		panic(err)
	}
}
func main() {

	linkURL := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true",
		user,
		pass,
		remoteHost,
		dbName,
	)
	db, err := sql.Open("mysql", linkURL)
	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		panic(err)
	}
	util.SetDB("main", db)
	defer db.Close()
	var totalGroup = NewQuestionGroup()

	fmt.Println("scan for categories")
	categories, err := FAQ.NewCategoryTree(APPID)
	if err != nil {
		panic(err)
	}
	_ = categories
	f, err := os.Create(volumePath + "/pprof")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	var size int

	err = db.QueryRow("SELECT count(Question_Id) FROM " + APPID + "_question WHERE Status != 0").Scan(&size)
	if err != nil {
		panic(err)
	}
	db.QueryRow("SELECT count(Question_Id) FROM " + APPID + "_question WHERE Status = 0").Scan(&size)

	rows, err := db.Query("SELECT Question_Id, CategoryId, Content  FROM " + APPID + "_question")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	fmt.Println("scan for stdQ")
	var count = 0

	for rows.Next() {
		count++
		fmt.Printf("\rOn %d/%d", count, size)

		var qId, catId int
		var content string
		err = rows.Scan(&qId, &catId, &content)
		if err != nil {
			panic(err)
		}
		var q = &StdQuestion{
			ID:      qId,
			CatID:   catId,
			Content: content,
			sq:      []*simQ{},
		}
		totalGroup.AddStdQ(q)
	}
	size = 0

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
	rows, err = db.Query("SELECT SQ_Id, Question_Id,Content FROM " + APPID + "_squestion ")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	fmt.Println("\r\nscan for SQ")
	count = 0
	for rows.Next() {
		var sqId, qId int
		var content string
		count++

		fmt.Printf("\rOn %d/%d", count, size)

		rows.Scan(&sqId, &qId, &content)
		var sq = &simQ{
			ID:      sqId,
			Content: content,
			stdQId:  qId,
		}
		totalGroup.AddSimQ(sq)
	}
	var badQuestions = []Question{}
	fmt.Printf("\nvalidation start\n")
	size = len(totalGroup.generalDict)
	count = 0
	for _, questions := range totalGroup.generalDict {
		count++
		fmt.Printf("\rProgess(%d/%d)...", count, size)
		var shouldTag bool = false
		for i, q := range questions {
			if q.Validate() {
				break
			}
			if i == len(questions)-1 {
				shouldTag = true
			}
		}
		if shouldTag {
			badQuestions = append(badQuestions, questions...)
		}
	}

	fmt.Printf("\n%d questions scanned\n", len(totalGroup.generalDict))
	var uqContent = make(map[string]bool, len(badQuestions))
	for _, q := range badQuestions {
		uqContent[q.ToContent()] = true
	}
	var output string
	for content := range uqContent {
		output += content + ";"
	}
	fmt.Printf("%d num of bad Question\n", len(uqContent))
	err = ioutil.WriteFile(volumePath+"/badcase", []byte(output), 0644)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("detail log are write to" + volumePath + "/badcase")
	fmt.Println("organize as stdQ")
	var syncQuestionID = make(map[int]struct{})
	for _, q := range badQuestions {
		switch v := q.(type) {
		case *StdQuestion:
			syncQuestionID[v.ID] = struct{}{}
		case *simQ:
			syncQuestionID[v.stdQId] = struct{}{}
		}
	}
	var input string
	fmt.Printf("\nDo you want to repair failed total [%d] of data?(Y/N)\n", len(syncQuestionID))
	reader := bufio.NewReader(os.Stdin)
	input, _ = reader.ReadString('\n')
	input = strings.Trim(input, "\n")
	if lower_input := strings.ToLower(input); lower_input == "y" {
		tx, _ := db.Begin()
		defer tx.Rollback()
		stmt, _ := tx.Prepare("UPDATE " + APPID + "_question SET status = 1 WHERE Question_Id = ?")
		defer stmt.Close()
		for qID := range syncQuestionID {
			_, err := stmt.Exec(qID)
			if err != nil {
				panic(err)
			}
		}
		tx.Commit()

		_, err := util.McManualBusiness(APPID)
		if err != nil {
			panic(err)
		}
		fmt.Println("MC Called sucess, finish repaired")
	} else if lower_input == "n" {
		fmt.Println("bye")
	} else {
		fmt.Println("unknow input " + lower_input)
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

type QuestionGroup struct {
	generalDict map[string][]Question
}

func NewQuestionGroup() *QuestionGroup {
	return &QuestionGroup{
		generalDict: make(map[string][]Question),
	}
}

//AddStdQ will add new StdQ into QuestionGroup's StdQDict, if collision is happened, old one will be overwrite.
func (qg *QuestionGroup) AddStdQ(q *StdQuestion) {
	qs, _ := qg.generalDict[q.Content]
	qs = append(qs, q)
	qg.generalDict[q.Content] = qs
}

func (qg *QuestionGroup) AddSimQ(q *simQ) {
	qs, _ := qg.generalDict[q.Content]
	qs = append(qs, q)
	qg.generalDict[q.Content] = qs
}

type simQ struct {
	ID      int
	Content string
	stdQId  int
	bad     bool
}
type StdQuestion struct {
	ID      int
	Content string
	CatID   int
	sq      []*simQ
	bad     bool
}

type Question interface {
	Validate() bool
	ToContent() string
}

//SolrQuery use for solr query api
func SolrQuery(host string, query string) (solrResponse, error) {
	req, err := http.NewRequest(http.MethodGet, "http://"+solrHost+":8081/solr/3rd_core/select", nil)
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

func (q *StdQuestion) Validate() bool {
	//Because STDQ have two id format, so it have to check twice...
	query := fmt.Sprintf("id:vipshop_other_%d_%d_self", q.ID, q.ID)
	resp, err := SolrQuery(remoteHost, query)
	if err != nil {
		panic(err)
	}
	if resp.Num == 1 {
		return true
	}
	query = fmt.Sprintf("id:vipshop_other_%d_self", q.ID, q.ID)
	resp, err = SolrQuery(remoteHost, query)
	if err != nil {
		panic(err)
	}
	return resp.Num == 1
}

func (sq *simQ) Validate() bool {
	query := fmt.Sprintf("id:vipshop_other_%d_%d", sq.stdQId, sq.ID)
	resp, err := SolrQuery(remoteHost, query)
	if err != nil {
		panic(err)
	}
	return resp.Num == 1
}

func (q *StdQuestion) ToContent() string {
	return q.Content
}

func (q *simQ) ToContent() string {
	return q.Content
}
