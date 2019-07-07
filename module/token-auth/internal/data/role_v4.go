package data

type RoleV4 struct {
	ID          int               `json:"-"`
	UUID        string            `json:"uuid"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Privileges  []string          `json:"privileges"`
	Menu        []*ModuleDetailV4 `json:"menu"`
	UserCount   int               `json:"user_count"`
}

func NewRoleV4() *RoleV4 {
	return &RoleV4{}
}
