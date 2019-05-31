package Service

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/pkg/logger"
)

var (
	tempFileName = ""
)

func setup() {
	logger.SetLevel("TRACE")
	tempFile, err := ioutil.TempFile("/tmp", "test")
	if err != nil {
		panic("Init temp file fail")
	}
	tempFile.WriteString(fmt.Sprintln("ADMIN_SERVICE_NLU=http://172.16.101.98:13901/"))
	tempFile.Close()
	tempFileName = tempFile.Name()
	err = util.LoadConfigFromFile(tempFile.Name())
	if err != nil {
		panic("Load config fail")
	}
}

func tearDown() {
	if tempFileName != "" {
		err := os.Remove(tempFileName)
		if err != nil {
			panic(err.Error())
		}
	}
}

func TestGetRecommandStdQuestion(t *testing.T) {
	setup()
	Init()
	ret, err := GetRecommandStdQuestion("csbot", "我", 3)
	if err != nil {
		t.Error(err.Error())
	} else {
		fmt.Printf("%+v\n", ret)
	}
}

func BenchmarkGetRecommandStdQuestion(b *testing.B) {
	setup()
	Init()
	GetRecommandStdQuestion("csbot", "我", 3)
	for i := 0; i < b.N; i++ {
		GetRecommandStdQuestion("csbot", "我", 3)
	}
}

//FIX ME: Need mock out of NLU dependency
// func TestNLUService(t *testing.T) {
// 	setup()
// 	testSentence := "你是男是女"
// 	appid := "csbot"
// 	result, err := GetNLUResult(appid, testSentence)
// 	if err != nil {
// 		t.Error(err.Error())
// 	}
// 	fmt.Printf("%#v\n", result)
// 	defer tearDown()
// }

//FIX ME: Need mock out of NLU dependency
// func TestNLUServices(t *testing.T) {
// 	setup()
// 	testSentence := []string{
// 		"你是男是女",
// 		"你会什么",
// 	}
// 	appid := "csbot"
// 	results, err := GetNLUResults(appid, testSentence)
// 	if err != nil {
// 		t.Error(err.Error())
// 	}

// 	for _, result := range results {
// 		fmt.Println("Orig question:", result.Sentence)
// 		fmt.Println(result.Segment.ToString())
// 		fmt.Println(result.Segment.ToFullString())
// 		fmt.Println(result.Keyword.ToString())
// 		fmt.Println(result.SentenceType)
// 		fmt.Println("===================")
// 	}
// 	defer tearDown()
// }
