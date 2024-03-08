package utils

const (
	ReqBodyKey = "req-body"
	ResBodyKey = "res-body"
)

type ResponseResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
}
