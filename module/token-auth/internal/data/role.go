package data

type Role struct {
	UUID        string              `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Privileges  map[string][]string `json:"privileges"`
	UserCount   int                 `json:"user_count"`
}

type Module struct {
	ID       int      `json:"-"`
	Code     string   `json:"code"`
	Name     string   `json:"name"`
	Commands []string `json:"commands"`
	Status   bool     `json:"-"`
}
