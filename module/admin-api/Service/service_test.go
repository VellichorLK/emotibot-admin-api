package Service

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"emotibot.com/emotigo/module/admin-api/util"
)

var (
	tempFileName = ""
)

func setup() {
	util.LogInit("TRACE")
	tempFile, err := ioutil.TempFile("/tmp", "test")
	if err != nil {
		panic("Init temp file fail")
	}
	tempFile.WriteString(fmt.Sprintln("SERVICE_NLU=http://172.16.101.98:13901/"))
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

func TestNLUService(t *testing.T) {
	setup()
	testSentence := "你是男是女"
	appid := "csbot"
	result, err := GetNLUResult(appid, testSentence)
	if err != nil {
		t.Error(err.Error())
	}
	fmt.Printf("%#v\n", result)
	defer tearDown()
}

func TestNLUServices(t *testing.T) {
	setup()
	testSentence := []string{
		"你是男是女",
		"你会什么",
	}
	appid := "csbot"
	results, err := GetNLUResults(appid, testSentence)
	if err != nil {
		t.Error(err.Error())
	}

	for _, result := range results {
		fmt.Println("Orig question:", result.Sentence)
		fmt.Println(result.Segment.ToString())
		fmt.Println(result.Segment.ToFullString())
		fmt.Println(result.Keyword.ToString())
		fmt.Println(result.SentenceType)
		fmt.Println("===================")
	}
	defer tearDown()
}
