//Package api used for api client
package api

import "net/http"

//HTTPClient should be able to delegated by others to do the http request.
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}
