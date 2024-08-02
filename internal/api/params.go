package api

type Params struct {
	Key string `param:"key" validate:"required"`
}
