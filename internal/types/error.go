package types

type v2ApiError struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

type v1ApiError struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
