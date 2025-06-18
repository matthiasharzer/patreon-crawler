package api

type UserErrorResponse struct {
	Errors []ResponseError `json:"errors"`
}

type ResponseError struct {
	CodeName string `json:"code_name"`
	Code     int    `json:"code"`
}
