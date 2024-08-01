package api

type BadRequest struct {
	Message string `enums:"key is expired" json:"message"`
}

type NotFound struct {
	Message string `enums:"key not exist" json:"message"`
}

type UnprocessableEntity struct {
	Message string `json:"message"`
}

type TooManyRequests struct {
	Message string `enums:"connection timeout,context timeout" json:"message"`
}

type InternalServer struct {
	Message string `json:"message"`
}
