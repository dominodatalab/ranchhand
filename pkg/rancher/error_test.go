package rancher

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthError(t *testing.T) {
	err := &authError{Body: "response body"}
	assert.Equal(t, "rancher login unauthorized: response body", err.Error())
}

func TestIsUnauthorized(t *testing.T) {
	testcases := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil", nil, false},
		{"random", errors.New("arbitrary"), false},
		{"autherr", &authError{}, true},
	}

	for _, c := range testcases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.expected, IsUnauthorized(c.err))
		})
	}
}
