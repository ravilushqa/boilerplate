package greeter_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ravilushqa/boilerplate/greeter"
)

func TestServiceGreet(t *testing.T) {
	t.Parallel()

	maxName := strings.Repeat("я", greeter.MaxNameLength)

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{name: "simple", input: "World", want: "Hello, World!"},
		{name: "trimmed", input: "  Gopher \n", want: "Hello, Gopher!"},
		{name: "unicode", input: "Мир", want: "Hello, Мир!"},
		{name: "inner spaces", input: "Rob Pike", want: "Hello, Rob Pike!"},
		{name: "max length", input: maxName, want: "Hello, " + maxName + "!"},
		{name: "empty", input: "", wantErr: greeter.ErrEmptyName},
		{name: "only spaces", input: "   ", wantErr: greeter.ErrEmptyName},
		{name: "too long", input: maxName + "a", wantErr: greeter.ErrNameTooLong},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var s greeter.Service

			got, err := s.Greet(t.Context(), tt.input)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
