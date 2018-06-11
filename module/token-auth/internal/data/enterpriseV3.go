package data

// Enterprise store the basic logging information of enterprise
// which can has multi app and user in it
type EnterpriseV3 struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type EnterpriseDetailV3 struct {
	EnterpriseV3
	Modules []*ModuleV3 `json:"modules"`
}

type EnterpriseAdminRequestV3 struct {
	Account 	string `json:"account"`
	Name		string `json:"name"`
	Password 	string `json:"password"`
}
