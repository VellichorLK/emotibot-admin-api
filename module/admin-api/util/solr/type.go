package solr

import (
	"encoding/json"
)

type QueryResponse struct {
	ResponseHeader ResponseHeader `json:"responseHeader"`
	Response       Response       `json:"response"`
}

type ResponseHeader struct {
	Status int64       `json:"status"`
	QTime  int64       `json:"QTime"`
	Params interface{} `json:"params"`
}

type Response struct {
	NumFound int64            `json:"numFound"`
	Start    int64            `json:"start"`
	Docs     *json.RawMessage `json:"docs"`
}
