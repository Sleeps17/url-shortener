package http_server

const (
	StatusOK    = "OK"
	StatusError = "Error"
)

const (
	InternalError         = "internal error"
	BadRequest            = "bad request"
	AliasAlreadyExist     = "alias already exist"
	AliasNotFound         = "alias not found"
	NewAliasAlreadyExists = "new_alias cannot use, url with this alias already exists"
)

const (
	Path = "http://localhost:8080/"
)
