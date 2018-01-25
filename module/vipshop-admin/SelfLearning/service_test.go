package SelfLearning

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"emotibot.com/emotigo/module/vipshop-admin/SelfLearning/internal/data"
	"emotibot.com/emotigo/module/vipshop-admin/util"
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
	util.LogInit(os.Stdout, os.Stdout, os.Stdout, os.Stdout)
	if ok := data.InitializeWord2Vec("../"); !ok {
		return fmt.Errorf("init failed")
	}

	cwd, _ := os.Getwd()
	filePath := cwd + "/../user_q.txt"
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
	util.LogInit(empty, empty, empty, empty)
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
