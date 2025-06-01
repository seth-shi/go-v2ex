package types

type V2ApiError struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

type V1ApiError struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (e V1ApiError) Success() bool {
	return e.Status == ""
}
