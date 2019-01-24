package solr

type AddCmd struct {
	Add AddCmdBody `json:"add"`
}

type AddCmdBody struct {
	Doc interface{} `json:"doc"`
}

type UpdateCmd struct {
	Add UpdateCmdBody `json:"add"`
}

type UpdateCmdBody struct {
	Doc interface{} `json:"doc"`
}

type DeleteByIDCmd struct {
	Delete DeleteByIDCmdBody `json:"delete"`
}

type DeleteByIDCmdBody struct {
	ID string `json:"id"`
}

type DeleteByQueryCmd struct {
	Delete DeleteByQueryCmdBody `json:"delete"`
}

type DeleteByQueryCmdBody struct {
	Query string `json:"query"`
}
