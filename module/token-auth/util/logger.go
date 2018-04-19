package util

import (
	"fmt"
	"io"
	"log"
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
)

// LogInit should be called before server start
func LogInit(
	traceHandle io.Writer,
	infoHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer, prefix string) {

	LogTrace = log.New(traceHandle,
		fmt.Sprintf("[%s] TRACE: ", prefix),
		log.Ldate|log.Ltime|log.Lshortfile)

	LogInfo = log.New(infoHandle,
		fmt.Sprintf("[%s] INFO: ", prefix),
		log.Ldate|log.Ltime|log.Lshortfile)

	LogWarn = log.New(warningHandle,
		fmt.Sprintf("[%s] WARNING: ", prefix),
		log.Ldate|log.Ltime|log.Lshortfile)

	LogError = log.New(errorHandle,
		fmt.Sprintf("[%s] ERROR: ", prefix),
		log.Ldate|log.Ltime|log.Lshortfile)
}
