package models

type ApiResponse struct {
	Code    int32 `json:"code"`
	Type    string `json:"type"`
	Message interface{}  `json:"message"`
}
