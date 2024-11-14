package errors

import "fmt"

func Chain(msg string, err error) error {
	return fmt.Errorf("ERROR: %s\n%w", msg, err)
}
