package greeter_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ravilushqa/boilerplate/internal/greeter"
)

func TestServiceGreet(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		input string
		want  string
		err   error
	}{
		"Simple":      {input: "World", want: "Hello, World!"},
		"Trimmed":     {input: "  Gopher \n", want: "Hello, Gopher!"},
		"Unicode":     {input: "Мир", want: "Hello, Мир!"},
		"Empty":       {input: "", err: greeter.ErrEmptyName},
		"OnlySpaces":  {input: "   ", err: greeter.ErrEmptyName},
		"MaxLength":   {input: strings.Repeat("я", greeter.MaxNameLength), want: "Hello, " + strings.Repeat("я", greeter.MaxNameLength) + "!"},
		"TooLong":     {input: strings.Repeat("a", greeter.MaxNameLength+1), err: greeter.ErrNameTooLong},
		"InnerSpaces": {input: "Rob Pike", want: "Hello, Rob Pike!"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := greeter.New().Greet(t.Context(), tc.input)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}
