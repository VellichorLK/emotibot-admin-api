package data

type RoleV3 struct {
	ID          int                 `json:"-"`
	UUID        string              `json:"uuid"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Privileges  map[string][]string `json:"privileges"`
	UserCount   int                 `json:"user_count"`
}

func NewRoleV3() *RoleV3 {
	return &RoleV3{
		Privileges: make(map[string][]string),
	}
}
