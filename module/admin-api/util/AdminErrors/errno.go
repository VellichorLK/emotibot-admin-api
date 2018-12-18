package AdminErrors

import "net/http"

const (
	// ErrnoSuccess is errno of success
	ErrnoSuccess = 0

	// ErrDBError is error about db like db disconnect, schema error, etc
	ErrnoDBError = -1

	// ErrnoIOError, like create dir fail, create file err, etc
	ErrnoIOError = -2

	// ErrnoRequestError, like error input
	ErrnoRequestError = -3

	// ErrnoConsulService means Consul service is not available
	ErrnoConsulService = -4

	// ErrnoOpenAPI means Openapi service is not available
	ErrnoOpenAPI = -5

	// ErrnoJsonParse means error when parse json string from other service
	ErrnoJSONParse = -6

	// ErrnoAPIError means error when calling remote API
	ErrnoAPIError = -7

	// ErrnoNotFound means resource not found
	ErrnoNotFound = -8

	// ErrnoBase64Decode means error when decode base64
	ErrnoBase64Decode = -9

	// ErrnoInitfailed means the package have not initialed resources for the api.
	ErrnoInitfailed = -10
)

func GetReturnStatus(errno int) int {
	switch errno {
	case ErrnoSuccess:
		return http.StatusOK
	case ErrnoNotFound:
		return http.StatusNotFound
	case ErrnoRequestError:
		return http.StatusBadRequest
	case ErrnoDBError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
