package cu

import (
	"bytes"
	"io/ioutil"
	"log"
	"testing"

	"emotibot.com/emotigo/pkg/logger"
)

func TestModuleInfoInitEmotionEngineClient(t *testing.T) {
	buffer := &bytes.Buffer{}
	warnBuffer := &bytes.Buffer{}
	logger.Error = log.New(buffer, "", 0)
	logger.Warn = log.New(warnBuffer, "", 0)
	callback, found := ModuleInfo.OneTimeFunc[keyInitEmotionEngine]
	if !found {
		t.Fatal("expect to found one time func in ModuleInfo")
	}
	callback()
	data, _ := ioutil.ReadAll(buffer)
	if len(data) == 0 {
		t.Fatal("expect it to failed, but it finished without printing error")
	}
	buffer.Reset()
	ModuleInfo.Environments = map[string]string{
		"EMOTION_ENGINE_URL": "http://localhost:8080",
	}
	callback()
	data, _ = ioutil.ReadAll(buffer)
	if len(data) > 0 {
		t.Fatal(string(data))
	}
	data, _ = ioutil.ReadAll(warnBuffer)
	if len(data) == 0 {
		t.Fatalf("warn log should log, but nothing log:\n%s", data)
	}
	if filterScore != 60 {
		t.Error("expect filterScore to be default 60, but got ", filterScore)
	}

	ModuleInfo.Environments["EMOTION_FILTER_SCORE"] = "80"
	callback()
	if filterScore != 80 {
		t.Error("expect score change to 80, but got", filterScore)
	}

}
