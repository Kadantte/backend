package errors

type _error struct{}

func (e *_error) Error() string { return "" }

var Err = &_error{}
