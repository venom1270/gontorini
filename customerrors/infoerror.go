package customerrors

// Custom error, meant for displaying erros to the user ("info") and not completly stopping the current logic/process
type InfoError struct {
	err string
}

// Error implements error.
func (e InfoError) Error() string {
	return e.err
}

func NewInfoError(s string) InfoError {
	return InfoError{s}
}
