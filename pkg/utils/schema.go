package utils

const (
	ResBodyKey = "res-body"
)

type ResponseResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
}
