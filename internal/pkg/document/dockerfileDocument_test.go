package document

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	testCases := []struct {
		name       string
		content    string
		newContent string
		result     bool
	}{
		{
			name:       "flag added",
			content:    "FROM scratch",
			newContent: "FROM --platform=abc scratch",
			result:     true,
		},
		{
			name:       "flag value changed",
			content:    "FROM --platform=abc scratch",
			newContent: "FROM --platform=abcd scratch",
			result:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc := NewDockerfileDocument("file:///tmp/Dockerfile", 1, []byte(tc.content))
			result := doc.Update(2, []byte(tc.newContent))
			require.Equal(t, tc.result, result)
		})
	}
}
