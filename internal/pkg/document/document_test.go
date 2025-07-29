package document

import (
	"context"
	"runtime"
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

func TestDocumentPath(t *testing.T) {
	testCases := []struct {
		name     string
		u        uri.URI
		folder   string
		fileName string
		wsl      bool
	}{
		{
			name:     "Linux URI",
			u:        "file:///tmp/Dockerfile",
			folder:   "/tmp",
			fileName: "Dockerfile",
			wsl:      false,
		},
		{
			name:     "Windows wsl$ host URI",
			u:        "file://wsl%24/docker-desktop/tmp/Dockerfile",
			folder:   "/docker-desktop/tmp",
			fileName: "Dockerfile",
			wsl:      true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mgr := NewDocumentManager()
			document := NewDocument(mgr, tc.u, protocol.DockerfileLanguage, 1, []byte{})
			path, err := document.DocumentPath()
			require.NoError(t, err)
			require.Equal(t, tc.folder, path.Folder)
			require.Equal(t, tc.fileName, path.FileName)
			require.Equal(t, tc.wsl, path.WSLDollarSignHost)
		})
	}
}

func TestDocumentPath_Windows(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.SkipNow()
		return
	}

	testCases := []struct {
		name     string
		u        uri.URI
		folder   string
		fileName string
		wsl      bool
	}{
		{
			name:     "Windows c%3A URI",
			u:        "file:///c%3A/tmp/Dockerfile",
			folder:   "c:\\tmp",
			fileName: "Dockerfile",
			wsl:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mgr := NewDocumentManager()
			document := NewDocument(mgr, tc.u, protocol.DockerfileLanguage, 1, []byte{})
			path, err := document.DocumentPath()
			require.NoError(t, err)
			require.Equal(t, tc.folder, path.Folder)
			require.Equal(t, tc.fileName, path.FileName)
			require.Equal(t, tc.wsl, path.WSLDollarSignHost)
		})
	}
}
