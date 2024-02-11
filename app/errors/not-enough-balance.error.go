package errors

type notEnoughBalanceError struct {
	message string
}

var NotEnoughBalanceError = &notEnoughBalanceError{}

func NewNotEnoughBalanceError() *notEnoughBalanceError {
	return &notEnoughBalanceError{
		message: "not enough balance",
	}
}

func (e *notEnoughBalanceError) Error() string {
	return e.message
}
