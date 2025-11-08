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
