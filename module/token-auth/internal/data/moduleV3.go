package data

type ModuleV3 struct {
	ID          int    `json:"-"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Status      bool   `json:"status"`
}

type ModuleDetailV3 struct {
	ModuleV3
	Description string `json:"description"`
	Commands []string `json:"commands"`
}
