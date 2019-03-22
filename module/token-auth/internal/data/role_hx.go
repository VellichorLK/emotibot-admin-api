package data

type RoleHX struct {
	ID          int                 `json:"id"`
	UUID        string              `json:"uuid"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Privileges  map[string][]string `json:"privileges"`
	UserCount   int                 `json:"user_count"`
}

func NewRoleHX() *RoleHX {
	return &RoleHX{
		Privileges: make(map[string][]string),
	}
}
