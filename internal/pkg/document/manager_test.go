package document

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.lsp.dev/uri"
)

func TestReadDocument(t *testing.T) {
	testFile := filepath.Join(os.TempDir(), "TestReadDocument")
	err := os.WriteFile(testFile, []byte("hello world"), 0644)
	require.NoError(t, err)

	contents, err := ReadDocument(uri.URI(fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(testFile), "/"))))
	require.NoError(t, err)
	require.Equal(t, "hello world", string(contents))
}

func TestURIfilename(t *testing.T) {
	var fn string
	var err error
	fn, err = filename(uri.URI("file:///mod"))
	require.NoError(t, err)
	assert.Equal(t, "/mod", fn)
	fn, err = filename(uri.URI("ext://mod"))
	require.Error(t, err)
	assert.Equal(t, "", fn)
	assert.Equal(t, "only file URIs are supported, got ext", err.Error())
}

func TestWrite(t *testing.T) {
	testCases := []struct {
		name       string
		content    string
		newContent string
		changed    bool
	}{
		{
			name:       "FROM instruction has changed",
			content:    "FROM alpine:3.16.1",
			newContent: "FROM alpine:3.16.2",
			changed:    true,
		},
		{
			name:       "whitespace added",
			content:    "FROM alpine:3.16.1",
			newContent: "FROM alpine:3.16.1 ",
			changed:    false,
		},
		{
			name:       "newline moves content",
			content:    "FROM alpine:3.16.1",
			newContent: "\nFROM alpine:3.16.1",
			changed:    true, // line numbers have changed
		},
		{
			name:       "newline at the end",
			content:    "FROM alpine:3.16.1 AS base",
			newContent: "FROM alpine:3.16.1 \\\nAS base",
			changed:    true, // FROM now spans two lines
		},
		{
			name:       "comments are not considered a change",
			content:    "FROM scratch",
			newContent: "FROM scratch\n# comment",
			changed:    false,
		},
		{
			name:       "syntax parser directive changed",
			content:    "#escape=\\\nFROM alpine:3.16.1 \\\nAS base",
			newContent: "#escape=`\nFROM alpine:3.16.1 \\\nAS base",
			changed:    true,
		},
		{
			name:       "check parser directive changed",
			content:    "#\nFROM scratch",
			newContent: "#check=skip=JSONArgsRecommended\nFROM scratch",
			changed:    true,
		},
		{
			name:       "add non check parser",
			content:    "FROM scratch",
			newContent: "FROM scratch\n#check=skip=JSONArgsRecommended",
			changed:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			manager := NewDocumentManager()
			defer manager.Remove("Dockerfile")

			changed, err := manager.Write(context.Background(), "Dockerfile", protocol.DockerfileLanguage, 1, []byte(tc.content))
			require.NoError(t, err)
			require.True(t, changed)
			version, err := manager.Version(context.Background(), "Dockerfile")
			require.NoError(t, err)
			require.Equal(t, int32(1), version)

			changed, err = manager.Write(context.Background(), "Dockerfile", protocol.DockerfileLanguage, 2, []byte(tc.newContent))
			require.NoError(t, err)
			require.Equal(t, tc.changed, changed)
			version, err = manager.Version(context.Background(), "Dockerfile")
			require.NoError(t, err)
			require.Equal(t, int32(2), version)
		})
	}
}
