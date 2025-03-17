package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConvertToHCLPosition(t *testing.T) {
	pos := ConvertToHCLPosition("", 0, 0)
	require.Equal(t, 1, pos.Line)
	require.Equal(t, 1, pos.Column)
	require.Equal(t, 0, pos.Byte)
}
