package dto

type SuccessMessageResponse struct {
	Message string `json:"message"`
}

type ListResponse struct {
	Payload    []any `json:"payload"`
	TotalPages int   `json:"total_pages"`
}

type ResponderOptions struct {
	Message string `json:"message"`
	Code    int
}
