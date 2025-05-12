package document

import (
	"context"
	"testing"

	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/stretchr/testify/require"
	"go.lsp.dev/uri"
)

func TestGetInstruction(t *testing.T) {
	testCases := []struct {
		name         string
		content      string
		positions    []protocol.Position
		found        []bool
		instructions []string
	}{
		{
			name:    "FROM alpine:3.16.1",
			content: "FROM alpine:3.16.1",
			positions: []protocol.Position{
				{Line: 0, Character: 0},
				{Line: 1, Character: 0},
			},
			found: []bool{
				true, false,
			},
			instructions: []string{"FROM", ""},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := WithReadDocumentFunc(func(uri.URI) ([]byte, error) {
				return []byte(tc.content), nil
			})
			mgr := NewDocumentManager(opts)
			doc, err := mgr.Read(context.Background(), "")
			require.Nil(t, err)

			for i := range tc.positions {
				instruction := doc.(DockerfileDocument).Instruction(tc.positions[i])
				if tc.found[i] {
					require.NotNil(t, instruction)
					require.Equal(t, tc.instructions[i], instruction.Value)
				} else {
					require.Nil(t, instruction)
				}
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	testCases := []struct {
		name             string
		content          string
		newContent       string
		hasNodes         bool
		instructionValue string
	}{
		{
			name:             "syntax error interrupts parsing",
			content:          "FROM alpine:3.16.1",
			newContent:       "FROM alpine:3.16.1\nRUN <<a",
			hasNodes:         false,
			instructionValue: "",
		},
		{
			name:             "instruction completely changed",
			content:          "FROM alpine:3.16.1",
			newContent:       "ARG imageTag",
			hasNodes:         true,
			instructionValue: "ARG",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			document := NewDocument(NewDocumentManager(), "", protocol.DockerfileLanguage, 1, []byte(tc.content))
			document.Update(2, []byte(tc.newContent))
			if tc.hasNodes {
				require.Equal(t, tc.instructionValue, document.(DockerfileDocument).Nodes()[0].Value)
			} else {
				require.Len(t, document.(DockerfileDocument).Nodes(), 0)
			}
		})
	}
}

func TestGetDecoder(t *testing.T) {
	testCases := []struct {
		name    string
		content string
	}{
		{
			name:    "target {}",
			content: "target {}",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := WithReadDocumentFunc(func(uri.URI) ([]byte, error) {
				return []byte(tc.content), nil
			})
			mgr := NewDocumentManager(opts)
			document, err := mgr.Read(context.Background(), "docker-bake.hcl")
			require.Nil(t, err)

			hclDocument, ok := document.(BakeHCLDocument)
			require.True(t, ok)
			require.NotNil(t, hclDocument.Decoder())
		})
	}
}
