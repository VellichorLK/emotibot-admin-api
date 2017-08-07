package auth

import (
	"io"
	"log"
)

var (
	LogTrace *log.Logger
	LogInfo  *log.Logger
	LogWarn  *log.Logger
	LogError *log.Logger
)

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

//func main() {
//	Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
//
//	Trace.Println("I have something standard to say")
//	Info.Println("Special Information")
//	Warning.Println("There is something you need to know about")
//	Error.Println("Something has failed")
//}
