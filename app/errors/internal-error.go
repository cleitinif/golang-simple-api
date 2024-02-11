package errors

type internalError struct {
	message string
}

var InternalError = &internalError{}

func NewInternalError() *internalError {
	return &internalError{
		message: "internal error",
	}
}

func (e *internalError) Error() string {
	return e.message
}
