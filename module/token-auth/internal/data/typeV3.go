package data

type LoginInfoV3 struct {
	Token string        `json:"token"`
	Info  *UserDetailV3 `json:"info"`
}
