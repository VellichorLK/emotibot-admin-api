package util

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

var (
	// LogTrace is for debug only
	LogTrace *log.Logger

	// LogInfo is for normal log
	LogInfo *log.Logger

	// LogWarn if for protential error
	LogWarn *log.Logger

	// LogError if for critical error
	LogError *log.Logger

	logLevel = map[string]int{
		"ERROR": 0,
		"WARN":  1,
		"INFO":  2,
		"TRACE": 3,
	}
	levelCount = 4
	logPrefix  = ""
)

// LogInit should be called before server start
func LogInit(
	prefix string,
	handler ...io.Writer) {
	logPrefix = prefix

	for len(handler) < levelCount {
		handler = append(handler, ioutil.Discard)
	}
	fmt.Printf("handlers: %+v\n", handler)

	LogTrace = log.New(handler[logLevel["TRACE"]],
		fmt.Sprintf("[%s] TRACE: ", prefix),
		log.Ldate|log.Ltime|log.Lshortfile)

	LogInfo = log.New(handler[logLevel["INFO"]],
		fmt.Sprintf("[%s] INFO: ", prefix),
		log.Ldate|log.Ltime|log.Lshortfile)

	LogWarn = log.New(handler[logLevel["WARN"]],
		fmt.Sprintf("[%s] WARNING: ", prefix),
		log.Ldate|log.Ltime|log.Lshortfile)

	LogError = log.New(handler[logLevel["ERROR"]],
		fmt.Sprintf("[%s] ERROR: ", prefix),
		log.Ldate|log.Ltime|log.Lshortfile)
}

func SetLogLevel(level string) {
	minShowLevel, ok := logLevel[level]
	if !ok {
		minShowLevel = 1
	}
	output := make([]io.Writer, levelCount)
	for idx := range output {
		if idx <= minShowLevel {
			output[idx] = os.Stdout
		} else {
			output[idx] = ioutil.Discard
		}
	}
	LogInit(logPrefix, output...)
}
