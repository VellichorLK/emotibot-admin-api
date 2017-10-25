package util

import (
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
	errorHandle io.Writer) {

	LogTrace = log.New(traceHandle,
		"TRACE: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	LogInfo = log.New(infoHandle,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	LogWarn = log.New(warningHandle,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	LogError = log.New(errorHandle,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}
