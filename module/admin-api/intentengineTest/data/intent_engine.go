package data

type IELoadModelResp struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

type IEUnloadModelResp struct {
	Status string `json:"status"`
}

type IEPredictResp struct {
	Status      string          `json:"status"`
	Predictions []*IEPrediction `json:"predictions"`
	Error       string          `json:"error"`
}

type IEPrediction struct {
	Label     string        `json:"label"`
	Score     int64         `json:"score"`
	OtherInfo []interface{} `json:"other_info"`
	Error     string        `json:"error"`
}
