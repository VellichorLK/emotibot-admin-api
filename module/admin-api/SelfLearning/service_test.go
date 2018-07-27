package SelfLearning

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"./internal/data"
	"emotibot.com/emotigo/module/admin-api/util"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		os.Exit(1)
	}
	retCode := m.Run()
	os.Exit(retCode)

}

//logs contains non-duplicated slice of testing data
var logs []string

//seed control how to populate the testing data
var seed = initSeed("seed")

//size control the size of the testing data
var size = initSize("array_size")

func initSeed(env string) (s int64) {
	var envSeed, exists = os.LookupEnv(env)

	if !exists {
		s = 1
	}
	for _, b := range []byte(envSeed) {
		s += int64(b)
	}
	return
}

func initSize(env string) (s int) {
	arraySize, exists := os.LookupEnv(env)
	s, err := strconv.Atoi(arraySize)
	if !exists || err != nil {
		s = 500
	}
	return
}

func setup() error {
	NluURL = "http://172.16.101.47:13901"
	MinSizeCluster = 10
	EarlyStopThreshold = 3
	ClusteringBatch = 20
	//empty := bytes.NewBuffer([]byte{})
	//util.LogInit(empty, empty, empty, empty)
	util.LogInit("TEST", os.Stdout, os.Stdout, os.Stdout, os.Stdout)
	if ok := data.InitializeWord2Vec("../"); !ok {
		return fmt.Errorf("init failed")
	}

	filePath := "./testdata/user_q.txt"
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	logs = strings.Split(string(data), "\n")
	return nil
}

func newArray(size int) []string {
	var results = make([]string, size)
	rand.Seed(seed)
	for i := 0; i < size; i++ {
		results[i] = logs[rand.Intn(len(logs))]
	}
	return results
}

func BenchmarkGetClusteringResult(b *testing.B) {
	if err := setup(); err != nil {
		b.Fatal(err)
	}

	feedbacks := newArray(b.N)
	var feedbackIDs = make([]uint64, b.N)
	var length float64
	for i := 0; i < len(feedbacks); i++ {
		feedbackIDs = append(feedbackIDs, uint64(i))
		length += float64(utf8.RuneCountInString(feedbacks[i]))
	}
	avgLength := length / float64(len(feedbacks))

	b.ResetTimer()
	getClusteringResult(feedbacks, feedbackIDs)
	fmt.Printf("字串平均長度:%f", avgLength)
}

func TestWholeFileGetClusteringResult(t *testing.T) {
	var feedbacks = logs
	var feedbackID []uint64

	rand.Seed(time.Now().UnixNano())

	var length float64
	for i := 0; i < len(feedbacks); i++ {
		feedbackID = append(feedbackID, uint64(i))
		length += float64(utf8.RuneCountInString(feedbacks[i]))
	}

	results := getClusteringResult(feedbacks, feedbackID)
	f, err := os.Create("./golden_result.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	fmt.Printf("number of cluster:%v, num of tag:%v\n", len(results.clusters), len(results.clusters[0].tags))
	var clusters = make([]struct {
		ID        int      `json:"clusterID"`
		Questions []string `json:"questions"`
		Tags      []string `json:"tags"`
		HitRate   float64  `json:"hitrate"`
	}, 0)
	for i, c := range results.clusters {
		var texts []string
		for _, id := range c.feedbackID {
			texts = append(texts, feedbacks[id])
			fmt.Printf("%s\n", feedbacks[id])
		}
		var cluster = struct {
			ID        int      `json:"clusterID"`
			Questions []string `json:"questions"`
			Tags      []string `json:"tags"`
			HitRate   float64  `json:"hitrate"`
		}{
			ID:        i,
			Questions: texts,
			Tags:      c.tags,
			HitRate:   hitRate(texts, c.tags),
		}
		clusters = append(clusters, cluster)
	}
	var totalHitRate float64
	for _, c := range clusters {
		totalHitRate += c.HitRate
	}
	avgHitRate := totalHitRate / float64(len(clusters))
	if avgHitRate <= 0.6 {
		t.Fatalf("expect hit rate above 0.6, but got %v", avgHitRate)
	}
	data, err := json.Marshal(clusters)
	if err != nil {
		t.Fatal(err)
	}
	num, err := fmt.Fprintf(f, "%.4f: %+v", avgHitRate, string(data))
	if num == 0 {
		t.Fatal("Nothing write to file")
	}
	if err != nil {
		t.Fatal(err)
	}

}
func TestGetClusteringResult(t *testing.T) {

	var feedbacks = newArray(size)
	var feedbackID []uint64

	var length float64
	for i := 0; i < len(feedbacks); i++ {
		feedbackID = append(feedbackID, uint64(i))
		length += float64(utf8.RuneCountInString(feedbacks[i]))
	}
	avgLength := length / float64(len(feedbacks))

	results := getClusteringResult(feedbacks, feedbackID)

	if results == nil {

		//fmt.Printf("%d:  %+v\n", len(feedbackID), feedbacks)
		t.Fatal("results is nil")
	}
	if clusterSize := len(results.clusters); clusterSize > 0 {
		fmt.Println(clusterSize)
		fmt.Printf("avg sentence length %f \n", avgLength)
	} else {
		t.Fatal("no clusters")
	}
}

func TestConcurrencyGetClusteringResult(t *testing.T) {
	empty := bytes.NewBuffer([]byte{})
	util.LogInit("TEST", empty, empty, empty, empty)
	var feedbacks = newArray(size)
	var feedbackID []uint64

	var length float64
	for i := 0; i < len(feedbacks); i++ {
		feedbackID = append(feedbackID, uint64(i))
		length += float64(utf8.RuneCountInString(feedbacks[i]))
	}
	avgLength := length / float64(len(feedbacks))
	fmt.Printf("每句平均長度:%f\n", avgLength)
	receiver := make(chan *clusteringResult)
	for i := 0; i <= 9; i++ {
		go func(output chan *clusteringResult) {
			result := getClusteringResult(feedbacks, feedbackID)
			receiver <- result
		}(receiver)
	}

	for i := 0; i <= 9; i++ {

		result := <-receiver
		if result == nil {
			t.Fatal("result should not be nil")
		} else if result.clusters == nil || len(result.clusters) == 0 {
			t.Fatal("result cluster should have at least one cluster")
		}
		fmt.Printf("result cluster size: %d\n", len(result.clusters))
	}
}

func hitRate(sentences []string, tags []string) float64 {
	var hit = 0
hits:
	for _, sentence := range sentences {

		for _, tag := range tags {
			if strings.Contains(sentence, tag) {
				hit++
				continue hits
			}

		}
	}
	return float64(hit) / float64(len(sentences))
}

func TestGetRecommend(t *testing.T) {

	testData := make(map[string]*rresponse)
	inputString := make([]string, 3)
	inputString[0] = "how can i do?"
	inputString[1] = "nothing you can do!"
	inputString[2] = "do you fully prepare"
	answers := make([]string, 5)
	answers[0] = "the hell is waiting for you."
	answers[1] = "nothing you can do!"
	answers[2] = "No. Loser!"
	answers[3] = "yes. well you will loose."
	answers[4] = "yes. you should wait for the end of world"

	r := &rresponse{}
	r.OtherInfo = &responseInfo{}
	r.OtherInfo.Custom = &responseCustom{}
	r.OtherInfo.Custom.RelatedQ = make([]*relatedQ, 0)
	q00 := &relatedQ{UserQ: inputString[0], StdQ: answers[0], Score: 99.91}
	q01 := &relatedQ{UserQ: inputString[0], StdQ: answers[1], Score: 91.4}
	r.OtherInfo.Custom.RelatedQ = append(r.OtherInfo.Custom.RelatedQ, q00)
	r.OtherInfo.Custom.RelatedQ = append(r.OtherInfo.Custom.RelatedQ, q01)
	testData[inputString[0]] = r

	r = &rresponse{}
	r.OtherInfo = &responseInfo{}
	r.OtherInfo.Custom = &responseCustom{}
	r.OtherInfo.Custom.RelatedQ = make([]*relatedQ, 0)
	q00 = &relatedQ{UserQ: inputString[1], StdQ: answers[0], Score: 21.91}
	q01 = &relatedQ{UserQ: inputString[1], StdQ: answers[4], Score: 33.4}
	r.OtherInfo.Custom.RelatedQ = append(r.OtherInfo.Custom.RelatedQ, q00)
	r.OtherInfo.Custom.RelatedQ = append(r.OtherInfo.Custom.RelatedQ, q01)
	testData[inputString[1]] = r

	r = &rresponse{}
	r.OtherInfo = &responseInfo{}
	r.OtherInfo.Custom = &responseCustom{}
	r.OtherInfo.Custom.RelatedQ = make([]*relatedQ, 0)
	q00 = &relatedQ{UserQ: inputString[2], StdQ: answers[0], Score: 33.91}
	q01 = &relatedQ{UserQ: inputString[2], StdQ: answers[2], Score: 87.87}
	q02 := &relatedQ{UserQ: inputString[2], StdQ: answers[3], Score: 61.87}
	r.OtherInfo.Custom.RelatedQ = append(r.OtherInfo.Custom.RelatedQ, q00)
	r.OtherInfo.Custom.RelatedQ = append(r.OtherInfo.Custom.RelatedQ, q01)
	r.OtherInfo.Custom.RelatedQ = append(r.OtherInfo.Custom.RelatedQ, q02)
	testData[inputString[2]] = r

	th := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		args := &responseArg{}
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = json.Unmarshal(b, args)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		resp, err := json.Marshal(testData[args.Text0])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-type:", "application/json")
		w.Write(resp)

	}
	ts := httptest.NewServer(http.HandlerFunc(th))
	defer ts.Close()
	responseURL = ts.URL
	db, mock, _ := sqlmock.New()
	util.SetDB("main", db)
	rows := sqlmock.NewRows([]string{"Question_Id", "Content"}).AddRow(1, answers[0]).AddRow(2, answers[1]).AddRow(3, answers[2]).AddRow(4, answers[3]).AddRow(5, answers[4])
	mock.ExpectQuery("select Question_Id,Content from vipshop_question where Content ").WithArgs(answers[0], answers[1], answers[2], answers[3], answers[4]).WillReturnRows(rows)

	recommends, err := getRecommend("csbot", inputString)
	if err != nil {
		t.Fatal(err)
	}

	if len(recommends) != len(answers) {
		t.Fatalf("recommend num is not equals.\n %v", recommends)
	}
	for i := 0; i < len(recommends); i++ {
		if recommends[i].Content != answers[i] && recommends[i].QID == i+1 {
			t.Fatalf("recommends are not correct.!\n recommends:%v\n,id:%v\nanswers:%v, id:%v\n", recommends, recommends[i].QID, answers, i+1)
		}
	}
}
