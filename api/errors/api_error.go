package errors

type ResponseError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}