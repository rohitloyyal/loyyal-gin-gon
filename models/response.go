package models

type Response struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Body    interface{} `json:"body"`
	Error   interface{} `json:"error"`
}
