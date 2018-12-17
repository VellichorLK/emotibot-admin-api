package cu

type apiFlowCreateBody struct {
	CreateTime int64  `json:"create_time"`
	FileName   string `json:"file_name"`
}

type apiFlowCreateResp struct {
	UUID string `json:"id"`
}

type daoFlowCreate struct {
	typ          int
	leftChannel  int
	rightChannel int
	enterprise   string
	callTime     int64
	uploadTime   int64
	updateTime   int64
	fileName     string
	uuid         string
	user         string
}
