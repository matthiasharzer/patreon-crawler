package api

type ErrorResponse struct {
	Errors []ResponseError `json:"errors"`
}

type ResponseError struct {
	ID       string `json:"id"`
	CodeName string `json:"code_name"`
	Code     int    `json:"code"`
	Title    string `json:"title"`
	Detail   string `json:"detail"`
	Status   string `json:"status"`
}
