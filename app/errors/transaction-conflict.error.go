package errors

type transactionConflictError struct {
	message string
}

var TransactionConflictError = &transactionConflictError{}

func NewTransactionConflictError() *transactionConflictError {
	return &transactionConflictError{
		message: "transaction conflict. please try again.",
	}
}

func (e *transactionConflictError) Error() string {
	return e.message
}
