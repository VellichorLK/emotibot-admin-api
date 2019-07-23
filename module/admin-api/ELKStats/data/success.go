package data


type SuccessStatusResponse struct {
	Status int `json:"status"`
	Message string `json:"message"`
}

func NewSuccessStatusResponse(code int, message string) SuccessStatusResponse {
	return SuccessStatusResponse{
		Status: code,
		Message: message,
	}
}
