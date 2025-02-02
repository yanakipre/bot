package httpopenaiclient

type OpenAIError struct {
	Err error
}

func (e OpenAIError) Error() string {
	return e.Err.Error()
}

func (e OpenAIError) Unwrap() error {
	return e.Err
}

func handleError(err error) error {
	if err == nil {
		return nil
	}
	return &OpenAIError{Err: err}
}
