package response

import (
	"fmt"
)

type V2ApiError struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

type V1ApiError struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (e V1ApiError) Error() string {
	return fmt.Sprintf("[%s]%s", e.Status, e.Message)
}

func (e V2ApiError) Error() string {
	return e.Message
}
