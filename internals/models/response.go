package models

type Response struct {
	Success bool
	Message string
}

type ResponseSucces struct {
	Success bool
	Message string
	Result  any
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
