package rancher

import "fmt"

type authError struct {
	Body string
}

func (e *authError) Error() string {
	return fmt.Sprintf("rancher login unauthorized: %s", e.Body)
}

func IsUnauthorized(err error) bool {
	_, ok := err.(*authError)
	return ok
}
