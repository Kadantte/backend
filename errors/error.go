package errors

type error struct{}

func (e *error) Error() string { return "" }

var Error *error = &error{}

var Done *error = &error{}
