package SelfLearning

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
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

var logs []string
var envSeed, _ = os.LookupEnv("seed")
var seed = initSeed(envSeed)

func initSeed(seedstr string) (s int64) {
	if seedstr == "" {
		s = 1
	}
	for _, b := range []byte(seedstr) {
		s += int64(b)
	}
	return
}

func setup() error {
	NluURL = "http://172.16.101.47:13901"
	MinSizeCluster = 10
	MaxNumToCluster = 10
	EarlyStopThreshold = 3
	ClusteringBatch = 50
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
func TestGetClusteringResult(t *testing.T) {
	arraySize, _ := os.LookupEnv("array_size")
	size, err := strconv.Atoi(arraySize)
	if err != nil {
		size = 500
	}
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
