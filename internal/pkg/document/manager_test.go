package document

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.lsp.dev/uri"
)

func TestReadLoadResolve(t *testing.T) {
	f := newFixture(t)
	require.NoError(t, os.Mkdir("exts", 0755))
	cwd, err := os.Getwd()
	require.NoError(t, err)
	extsPath := filepath.Join(cwd, "exts")
	WithReadDocumentFunc(func(u uri.URI) ([]byte, error) {
		file := u.Filename()
		return ReadDocument(uri.File(file))
	})(f.m)

	hello := filepath.Join(extsPath, "hello")
	require.NoError(t, os.WriteFile(hello, []byte(`hello = lambda: print('Hi')`), 0644))
	require.NoError(t, os.WriteFile("doc", []byte("load('ext://hello', 'hello')\nhello()"), 0644))
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

type fixture struct {
	ctx context.Context
	m   *Manager
}

func newFixture(t *testing.T) *fixture {
	wd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(wd) })
	dir := t.TempDir()
	require.NoError(t, os.Chdir(dir))
	return &fixture{
		ctx: context.Background(),
		m:   NewDocumentManager(),
	}
}
