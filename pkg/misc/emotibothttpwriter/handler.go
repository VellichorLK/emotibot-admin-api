package emotibothttpwriter

import "net/http"

// EmotibotHTTPWriter is a custom http writer
type EmotibotHTTPWriter struct {
	origWriter http.ResponseWriter
	statusCode int
}

// New will return a custom http writer with input origin writer
func New(OrigWriter http.ResponseWriter) *EmotibotHTTPWriter {
	return &EmotibotHTTPWriter{
		origWriter: OrigWriter,
		statusCode: http.StatusOK,
	}
}

// Header will just do the same with origin Header
func (w EmotibotHTTPWriter) Header() http.Header {
	return w.origWriter.Header()
}

// Write will just do the same with origin Write
func (w EmotibotHTTPWriter) Write(content []byte) (int, error) {
	return w.origWriter.Write(content)
}

// WriteHeader will save status to local variable and can get via function
// after finish write status to origin http writer
func (w EmotibotHTTPWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.origWriter.WriteHeader(statusCode)
}

// GetStatusCode will get status from local variable, which cannot get from origin writer
func (w EmotibotHTTPWriter) GetStatusCode() int {
	return w.statusCode
}
