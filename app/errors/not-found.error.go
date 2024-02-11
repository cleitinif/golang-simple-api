package errors

type notFoundError struct {
	message string
}

var NotFoundError = &notFoundError{}

func NewNotFoundError() *notFoundError {
	return &notFoundError{
		message: "not found",
	}
}

func (e *notFoundError) Error() string {
	return "entity not found"
}
