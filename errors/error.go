package errors

type _error struct{}

func (e *_error) Error() string { return "" }

var Error *_error = &_error{}
