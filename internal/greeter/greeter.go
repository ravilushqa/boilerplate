// Package greeter is an example of business logic.
//
// It knows nothing about HTTP or gRPC: transports call it and translate
// its errors into protocol-specific responses. Replace it with your own
// domain packages, keeping the same shape.
package greeter

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"
)

// MaxNameLength is the maximum accepted name length, in characters.
const MaxNameLength = 100

// Errors returned by [Service]. Transports map them with errors.Is.
var (
	ErrEmptyName   = errors.New("name must not be empty")
	ErrNameTooLong = errors.New("name is too long")
)

// Service greets people. The zero value is ready to use.
//
// A real service takes its dependencies (storage, clients) as fields
// set by a NewService constructor.
type Service struct{}

// Greet returns a greeting for name.
func (s *Service) Greet(_ context.Context, name string) (string, error) {
	name = strings.TrimSpace(name)

	switch {
	case name == "":
		return "", ErrEmptyName
	case utf8.RuneCountInString(name) > MaxNameLength:
		return "", ErrNameTooLong
	}

	return "Hello, " + name + "!", nil
}
