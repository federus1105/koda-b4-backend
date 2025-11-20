package models

type Response struct {
	Success bool `json:"success"`
	Message string `json:"message"`
}

type ResponseSucces struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Result  any    `json:"result"`
}

type PaginatedResponse[T any] struct {
	Page       int     `json:"page"`
	Limit      int     `json:"limit"`
	Total      int64   `json:"total"`
	TotalPages int     `json:"totalPages"`
	PrevURL    *string `json:"prevUrl"`
	NextURL    *string `json:"nextUrl"`
	Result     []T     `json:"data"`
}
