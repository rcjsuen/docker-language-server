package hcl

import (
	"context"
	"fmt"
	"testing"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/stretchr/testify/require"
)

func TestSemanticTokensFull(t *testing.T) {
	testCases := []struct {
		name    string
		content string
		result  [][]uint32
	}{
		{
			name:    "empty file",
			content: "",
			result:  [][]uint32{},
		},
		{
			name:    "target keyword only",
			content: "target{}",
			result:  [][]uint32{{0, 0, 6, SemanticTokenTypeIndex(TokenType_Type), 0}},
		},
		{
			name:    "target with name",
			content: "target \"api\" {}",
			result: [][]uint32{
				{0, 0, 6, SemanticTokenTypeIndex(TokenType_Type), 0},
				{0, 7, 5, SemanticTokenTypeIndex(TokenType_Class), 0},
			},
		},
		{
			name:    "target block without unindented target attribute",
			content: "target \"api\" {\ntarget = \"api\"\n}",
			result: [][]uint32{
				{0, 0, 6, SemanticTokenTypeIndex(TokenType_Type), 0},
				{0, 7, 5, SemanticTokenTypeIndex(TokenType_Class), 0},
				{1, 0, 6, SemanticTokenTypeIndex(TokenType_Property), 0},
				{0, 9, 5, SemanticTokenTypeIndex(TokenType_String), 0},
			},
		},
		{
			name:    "target block without indented target attribute",
			content: "target \"api\" {\n  target = \"api\"\n}",
			result: [][]uint32{
				{0, 0, 6, SemanticTokenTypeIndex(TokenType_Type), 0},
				{0, 7, 5, SemanticTokenTypeIndex(TokenType_Class), 0},
				{1, 2, 6, SemanticTokenTypeIndex(TokenType_Property), 0},
				{0, 9, 5, SemanticTokenTypeIndex(TokenType_String), 0},
			},
		},
		{
			name:    "# comment",
			content: "# comment",
			result: [][]uint32{
				{0, 0, 9, SemanticTokenTypeIndex(TokenType_Comment), 0},
			},
		},
		{
			name:    "# comment with a newline",
			content: "# comment\n",
			result: [][]uint32{
				{0, 0, 9, SemanticTokenTypeIndex(TokenType_Comment), 0},
			},
		},
		{
			name:    "# comment with a CR newline",
			content: "# comment\r\n",
			result: [][]uint32{
				{0, 0, 9, SemanticTokenTypeIndex(TokenType_Comment), 0},
			},
		},
		{
			name:    "# comment with newline and offset",
			content: " # comment\n",
			result: [][]uint32{
				{0, 1, 9, SemanticTokenTypeIndex(TokenType_Comment), 0},
			},
		},
		{
			name:    "// comment",
			content: "// comment",
			result: [][]uint32{
				{0, 0, 10, SemanticTokenTypeIndex(TokenType_Comment), 0},
			},
		},
		{
			name:    "// comment with a newline",
			content: "// comment\n",
			result: [][]uint32{
				{0, 0, 10, SemanticTokenTypeIndex(TokenType_Comment), 0},
			},
		},
		{
			name:    "// comment with CRLF",
			content: "// comment\r\n",
			result: [][]uint32{
				{0, 0, 10, SemanticTokenTypeIndex(TokenType_Comment), 0},
			},
		},
		{
			name:    "/* comment */",
			content: "/* comment */",
			result: [][]uint32{
				{0, 0, 13, SemanticTokenTypeIndex(TokenType_Comment), 0},
			},
		},
		{
			name:    "multiline /* comment */",
			content: "/* comment \n*/",
			result: [][]uint32{
				{0, 0, 11, SemanticTokenTypeIndex(TokenType_Comment), 0},
				{1, 0, 2, SemanticTokenTypeIndex(TokenType_Comment), 0},
			},
		},
		{
			name:    "multiline /* comment */ with newline and empty spaces",
			content: "/* comment \n\n*/",
			result: [][]uint32{
				{0, 0, 11, SemanticTokenTypeIndex(TokenType_Comment), 0},
				{2, 0, 2, SemanticTokenTypeIndex(TokenType_Comment), 0},
			},
		},
		{
			name:    "multiline /* comment */ with CRLF and empty spaces",
			content: "/* comment \r\n\r\n*/",
			result: [][]uint32{
				{0, 0, 11, SemanticTokenTypeIndex(TokenType_Comment), 0},
				{2, 0, 2, SemanticTokenTypeIndex(TokenType_Comment), 0},
			},
		},
		{
			name:    "multiline /* comment */ with CRLF",
			content: "/* comment \r\n*/",
			result: [][]uint32{
				{0, 0, 11, SemanticTokenTypeIndex(TokenType_Comment), 0},
				{1, 0, 2, SemanticTokenTypeIndex(TokenType_Comment), 0},
			},
		},
		{
			name:    "multiline /* comment */ with trailing newline",
			content: "/* comment \n*/\n",
			result: [][]uint32{
				{0, 0, 11, SemanticTokenTypeIndex(TokenType_Comment), 0},
				{1, 0, 2, SemanticTokenTypeIndex(TokenType_Comment), 0},
			},
		},
		{
			name:    "embedded /* comment",
			content: "target /* comment */ \"api\" {}",
			result: [][]uint32{
				{0, 0, 6, SemanticTokenTypeIndex(TokenType_Type), 0},
				{0, 7, 13, SemanticTokenTypeIndex(TokenType_Comment), 0},
				{0, 14, 5, SemanticTokenTypeIndex(TokenType_Class), 0},
			},
		},
		{
			name:    "target with name and comment",
			content: "# comment\ntarget \"api\" {}",
			result: [][]uint32{
				{0, 0, 9, SemanticTokenTypeIndex(TokenType_Comment), 0},
				{1, 0, 6, SemanticTokenTypeIndex(TokenType_Type), 0},
				{0, 7, 5, SemanticTokenTypeIndex(TokenType_Class), 0},
			},
		},
		{
			name:    "target with name two single-line comments",
			content: "# comment\n# comment\ntarget \"api\" {}",
			result: [][]uint32{
				{0, 0, 9, SemanticTokenTypeIndex(TokenType_Comment), 0},
				{1, 0, 9, SemanticTokenTypeIndex(TokenType_Comment), 0},
				{1, 0, 6, SemanticTokenTypeIndex(TokenType_Type), 0},
				{0, 7, 5, SemanticTokenTypeIndex(TokenType_Class), 0},
			},
		},
		{
			name:    "multiline comment after block",
			content: "group {\n}\n/*\n*/",
			result: [][]uint32{
				{0, 0, 5, SemanticTokenTypeIndex(TokenType_Type), 0},
				{2, 0, 2, SemanticTokenTypeIndex(TokenType_Comment), 0},
				{1, 0, 2, SemanticTokenTypeIndex(TokenType_Comment), 0},
			},
		},
		{
			name:    "bool variable ",
			content: "variable b {\n  default = true\n}",
			result: [][]uint32{
				{0, 0, 8, SemanticTokenTypeIndex(TokenType_Type), 0},
				{0, 9, 1, SemanticTokenTypeIndex(TokenType_Class), 0},
				{1, 2, 7, SemanticTokenTypeIndex(TokenType_Property), 0},
				{0, 10, 4, SemanticTokenTypeIndex(TokenType_Keyword), 0},
			},
		},
		{
			name:    "number variable ",
			content: "variable n {\n  default = 123\n}",
			result: [][]uint32{
				{0, 0, 8, SemanticTokenTypeIndex(TokenType_Type), 0},
				{0, 9, 1, SemanticTokenTypeIndex(TokenType_Class), 0},
				{1, 2, 7, SemanticTokenTypeIndex(TokenType_Property), 0},
				{0, 10, 3, SemanticTokenTypeIndex(TokenType_Number), 0},
			},
		},
		{
			name:    "null variable",
			content: "variable b {\n  default = null\n}",
			result: [][]uint32{
				{0, 0, 8, SemanticTokenTypeIndex(TokenType_Type), 0},
				{0, 9, 1, SemanticTokenTypeIndex(TokenType_Class), 0},
				{1, 2, 7, SemanticTokenTypeIndex(TokenType_Property), 0},
				{0, 10, 4, SemanticTokenTypeIndex(TokenType_Keyword), 0},
			},
		},
		{
			name:    "reference variable",
			content: "variable t { default = \"abc\" }\ntarget \"api\" { target = abc }",
			result: [][]uint32{
				{0, 0, 8, SemanticTokenTypeIndex(TokenType_Type), 0},
				{0, 9, 1, SemanticTokenTypeIndex(TokenType_Class), 0},
				{0, 4, 7, SemanticTokenTypeIndex(TokenType_Property), 0},
				{0, 10, 5, SemanticTokenTypeIndex(TokenType_String), 0},
				{1, 0, 6, SemanticTokenTypeIndex(TokenType_Type), 0},
				{0, 7, 5, SemanticTokenTypeIndex(TokenType_Class), 0},
				{0, 8, 6, SemanticTokenTypeIndex(TokenType_Property), 0},
				{0, 9, 3, SemanticTokenTypeIndex(TokenType_Variable), 0},
			},
		},
		{
			name:    "reference variable in the args attribute",
			content: "variable v { default = \"abc\" }\ntarget \"api\" { args = { var = v } }",
			result: [][]uint32{
				{0, 0, 8, SemanticTokenTypeIndex(TokenType_Type), 0},
				{0, 9, 1, SemanticTokenTypeIndex(TokenType_Class), 0},
				{0, 4, 7, SemanticTokenTypeIndex(TokenType_Property), 0},
				{0, 10, 5, SemanticTokenTypeIndex(TokenType_String), 0},
				{1, 0, 6, SemanticTokenTypeIndex(TokenType_Type), 0},
				{0, 7, 5, SemanticTokenTypeIndex(TokenType_Class), 0},
				{0, 8, 4, SemanticTokenTypeIndex(TokenType_Property), 0},
				{0, 9, 3, SemanticTokenTypeIndex(TokenType_Property), 0},
				{0, 6, 1, SemanticTokenTypeIndex(TokenType_Variable), 0},
			},
		},
		{
			name:    "target block with a string array",
			content: "target \"api\" {\n  tags = [\"tag\"] }",
			result: [][]uint32{
				{0, 0, 6, SemanticTokenTypeIndex(TokenType_Type), 0},
				{0, 7, 5, SemanticTokenTypeIndex(TokenType_Class), 0},
				{1, 2, 4, SemanticTokenTypeIndex(TokenType_Property), 0},
				{0, 8, 5, SemanticTokenTypeIndex(TokenType_String), 0},
			},
		},
		{
			name:    "variable block with a string array",
			content: "variable list {\n  default = [\"abc\"] }",
			result: [][]uint32{
				{0, 0, 8, SemanticTokenTypeIndex(TokenType_Type), 0},
				{0, 9, 4, SemanticTokenTypeIndex(TokenType_Class), 0},
				{1, 2, 7, SemanticTokenTypeIndex(TokenType_Property), 0},
				{0, 11, 5, SemanticTokenTypeIndex(TokenType_String), 0},
			},
		},
		{
			name:    "variable block within a string",
			content: "target api {\n  target = \"${var}\"\n}",
			result: [][]uint32{
				{0, 0, 6, SemanticTokenTypeIndex(TokenType_Type), 0},
				{0, 7, 3, SemanticTokenTypeIndex(TokenType_Class), 0},
				{1, 2, 6, SemanticTokenTypeIndex(TokenType_Property), 0},
				{0, 9, 1, SemanticTokenTypeIndex(TokenType_String), 0},
				{0, 1, 2, SemanticTokenTypeIndex(TokenType_Operator), 0},
				{0, 2, 3, SemanticTokenTypeIndex(TokenType_Variable), 0},
				{0, 3, 1, SemanticTokenTypeIndex(TokenType_Operator), 0},
				{0, 1, 1, SemanticTokenTypeIndex(TokenType_String), 0},
			},
		},
		{
			name:    "variable block within a complex string",
			content: "target api {\n  target = \"${var}string\"\n}",
			result: [][]uint32{
				{0, 0, 6, SemanticTokenTypeIndex(TokenType_Type), 0},
				{0, 7, 3, SemanticTokenTypeIndex(TokenType_Class), 0},
				{1, 2, 6, SemanticTokenTypeIndex(TokenType_Property), 0},
				{0, 9, 1, SemanticTokenTypeIndex(TokenType_String), 0},
				{0, 1, 2, SemanticTokenTypeIndex(TokenType_Operator), 0},
				{0, 2, 3, SemanticTokenTypeIndex(TokenType_Variable), 0},
				{0, 3, 1, SemanticTokenTypeIndex(TokenType_Operator), 0},
				{0, 1, 6, SemanticTokenTypeIndex(TokenType_String), 0},
				{0, 6, 1, SemanticTokenTypeIndex(TokenType_String), 0},
			},
		},
		{
			name:    "variable reference within an annotations array",
			content: "target api {\n  annotations = [ \"${var}string\" ]\n}",
			result: [][]uint32{
				{0, 0, 6, SemanticTokenTypeIndex(TokenType_Type), 0},
				{0, 7, 3, SemanticTokenTypeIndex(TokenType_Class), 0},
				{1, 2, 11, SemanticTokenTypeIndex(TokenType_Property), 0},
				{0, 16, 1, SemanticTokenTypeIndex(TokenType_String), 0},
				{0, 1, 2, SemanticTokenTypeIndex(TokenType_Operator), 0},
				{0, 2, 3, SemanticTokenTypeIndex(TokenType_Variable), 0},
				{0, 3, 1, SemanticTokenTypeIndex(TokenType_Operator), 0},
				{0, 1, 6, SemanticTokenTypeIndex(TokenType_String), 0},
				{0, 6, 1, SemanticTokenTypeIndex(TokenType_String), 0},
			},
		},
		{
			name:    "variable reference within a secret array",
			content: "target api {\n  secret = [ \"${var}string\" ]\n}",
			result: [][]uint32{
				{0, 0, 6, SemanticTokenTypeIndex(TokenType_Type), 0},
				{0, 7, 3, SemanticTokenTypeIndex(TokenType_Class), 0},
				{1, 2, 6, SemanticTokenTypeIndex(TokenType_Property), 0},
				{0, 11, 1, SemanticTokenTypeIndex(TokenType_String), 0},
				{0, 1, 2, SemanticTokenTypeIndex(TokenType_Operator), 0},
				{0, 2, 3, SemanticTokenTypeIndex(TokenType_Variable), 0},
				{0, 3, 1, SemanticTokenTypeIndex(TokenType_Operator), 0},
				{0, 1, 6, SemanticTokenTypeIndex(TokenType_String), 0},
				{0, 6, 1, SemanticTokenTypeIndex(TokenType_String), 0},
			},
		},
		{
			name:    "variable reference within a ssh array",
			content: "target api {\n  ssh = [ \"${var}string\" ]\n}",
			result: [][]uint32{
				{0, 0, 6, SemanticTokenTypeIndex(TokenType_Type), 0},
				{0, 7, 3, SemanticTokenTypeIndex(TokenType_Class), 0},
				{1, 2, 3, SemanticTokenTypeIndex(TokenType_Property), 0},
				{0, 8, 1, SemanticTokenTypeIndex(TokenType_String), 0},
				{0, 1, 2, SemanticTokenTypeIndex(TokenType_Operator), 0},
				{0, 2, 3, SemanticTokenTypeIndex(TokenType_Variable), 0},
				{0, 3, 1, SemanticTokenTypeIndex(TokenType_Operator), 0},
				{0, 1, 6, SemanticTokenTypeIndex(TokenType_String), 0},
				{0, 6, 1, SemanticTokenTypeIndex(TokenType_String), 0},
			},
		},
		{
			name:    "variable reference within a tags array",
			content: "target api {\n  tags = [ \"${var}string\" ]\n}",
			result: [][]uint32{
				{0, 0, 6, SemanticTokenTypeIndex(TokenType_Type), 0},
				{0, 7, 3, SemanticTokenTypeIndex(TokenType_Class), 0},
				{1, 2, 4, SemanticTokenTypeIndex(TokenType_Property), 0},
				{0, 9, 1, SemanticTokenTypeIndex(TokenType_String), 0},
				{0, 1, 2, SemanticTokenTypeIndex(TokenType_Operator), 0},
				{0, 2, 3, SemanticTokenTypeIndex(TokenType_Variable), 0},
				{0, 3, 1, SemanticTokenTypeIndex(TokenType_Operator), 0},
				{0, 1, 6, SemanticTokenTypeIndex(TokenType_String), 0},
				{0, 6, 1, SemanticTokenTypeIndex(TokenType_String), 0},
			},
		},
		{
			name:    "conditional expression by comparing strings, true expression has a string array",
			content: "target api {\n  platforms = \"\" == \"\" ? [ \"windows/arm64\" ] : []\n}",
			result: [][]uint32{
				{0, 0, 6, SemanticTokenTypeIndex(TokenType_Type), 0},
				{0, 7, 3, SemanticTokenTypeIndex(TokenType_Class), 0},
				{1, 2, 9, SemanticTokenTypeIndex(TokenType_Property), 0},
				{0, 12, 2, SemanticTokenTypeIndex(TokenType_String), 0},
				{0, 6, 2, SemanticTokenTypeIndex(TokenType_String), 0},
				{0, 7, 15, SemanticTokenTypeIndex(TokenType_String), 0},
			},
		},
		{
			name:    "conditional expression by comparing strings, false expression has a string array",
			content: "target api {\n  platforms = \"\" == \"\" ? [] : [ \"windows/arm64\" ]\n}",
			result: [][]uint32{
				{0, 0, 6, SemanticTokenTypeIndex(TokenType_Type), 0},
				{0, 7, 3, SemanticTokenTypeIndex(TokenType_Class), 0},
				{1, 2, 9, SemanticTokenTypeIndex(TokenType_Property), 0},
				{0, 12, 2, SemanticTokenTypeIndex(TokenType_String), 0},
				{0, 6, 2, SemanticTokenTypeIndex(TokenType_String), 0},
				{0, 12, 15, SemanticTokenTypeIndex(TokenType_String), 0},
			},
		},
		{
			name:    "conditional expression by comparing strings, true and false as the result",
			content: "target api {\n  pull = \"\" == \"\" ? true : false",
			result: [][]uint32{
				{0, 0, 6, SemanticTokenTypeIndex(TokenType_Type), 0},
				{0, 7, 3, SemanticTokenTypeIndex(TokenType_Class), 0},
				{1, 2, 4, SemanticTokenTypeIndex(TokenType_Property), 0},
				{0, 7, 2, SemanticTokenTypeIndex(TokenType_String), 0},
				{0, 6, 2, SemanticTokenTypeIndex(TokenType_String), 0},
				{0, 5, 4, SemanticTokenTypeIndex(TokenType_Keyword), 0},
				{0, 7, 5, SemanticTokenTypeIndex(TokenType_Keyword), 0},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc := document.NewBakeHCLDocument("", 1, []byte(tc.content))
			tokens, err := SemanticTokensFull(context.Background(), doc, "")
			require.NoError(t, err)
			testResultOffset := 0
			for i := 0; i < len(tokens.Data); i += 5 {
				if testResultOffset == len(tc.result) {
					break
				}

				// compare them as ints to make the differences more readable when a test fails
				require.Equal(t, int(tc.result[testResultOffset][0]), int(tokens.Data[i]), fmt.Sprintf("deltaLine mismatch for token index %v (token group %v)", i, i/5))
				require.Equal(t, int(tc.result[testResultOffset][1]), int(tokens.Data[i+1]), fmt.Sprintf("deltaStart mismatch for token index %v (token group %v)", i+1, i/5))
				require.Equal(t, int(tc.result[testResultOffset][2]), int(tokens.Data[i+2]), fmt.Sprintf("length mismatch for token index %v (token group %v)", i+2, i/5))
				require.Equal(t, int(tc.result[testResultOffset][3]), int(tokens.Data[i+3]), fmt.Sprintf("tokenType mismatch for token index %v (token group %v)", i+3, i/5))
				require.Equal(t, int(tc.result[testResultOffset][4]), int(tokens.Data[i+4]), fmt.Sprintf("tokenModifiers mismatch for token index %v (token group %v)", i+4, i/5))
				testResultOffset++
			}
			require.Len(t, tokens.Data, 5*len(tc.result))
		})
	}
}
