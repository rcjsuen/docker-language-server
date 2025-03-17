package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGitRepository(t *testing.T) {
	testCases := []struct {
		url      string
		expected string
	}{
		{url: "ssh://user@host.xz/path/to/repo.git/"},
		{url: "ssh://host.xz/path/to/repo.git/"},
		{url: "ssh://user@host.xz/path/to/repo.git/"},
		{url: "ssh://host.xz/path/to/repo.git/"},
		{url: "rsync://host.xz/path/to/repo.git/"},
		{url: "git://host.xz/path/to/repo.git/"},
		{url: "http://host.xz/path/to/repo.git/"},
		{url: "https://host.xz/path/to/repo.git/"},
		{
			url:      "ssh://user@host.xz:8888/path/to/repo.git/",
			expected: "host.xz:8888/path/to/repo.git",
		},
		{
			url:      "user@host.xz:path/to/repo.git",
			expected: "host.xz/path/to/repo.git",
		},
		{
			url:      "host.xz:/path/to/repo.git/",
			expected: "host.xz/path/to/repo.git",
		},
		{
			url:      "host.xz:/path/to/repo.git/",
			expected: "host.xz/path/to/repo.git",
		},
		{
			url:      "host.xz:path/to/repo.git",
			expected: "host.xz/path/to/repo.git",
		},
		{
			url:      "ssh://host.xz:8888/path/to/repo.git/",
			expected: "host.xz:8888/path/to/repo.git",
		},
		{
			url: "user@host.xz:/path/to/repo.git/",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.url, func(t *testing.T) {
			repository := GitRepository(tc.url)
			if tc.expected == "" {
				require.Equal(t, "host.xz/path/to/repo.git", repository)
			} else {
				require.Equal(t, tc.expected, repository)
			}
		})
	}
}

func TestWorkspaceFolders(t *testing.T) {
	testCases := []struct {
		name             string
		uri              string
		workspaceFolders []string
		folder           string
		absolutePath     string
		relativePath     string
	}{
		{
			name:             "simple prefix, folder without trailing slash",
			uri:              "file:///a/b/c/Dockerfile",
			workspaceFolders: []string{"/a/b/c"},
			folder:           "/a/b/c",
			absolutePath:     "/a/b/c/Dockerfile",
			relativePath:     "Dockerfile",
		},
		{
			name:             "simple prefix, folder with trailing slash",
			uri:              "file:///a/b/c/Dockerfile",
			workspaceFolders: []string{"/a/b/c/"},
			folder:           "/a/b/c/",
			absolutePath:     "/a/b/c/Dockerfile",
			relativePath:     "Dockerfile",
		},
		{
			name:             "shared prefix",
			uri:              "file:///a/b/c2/Dockerfile",
			workspaceFolders: []string{"/a/b/c", "/a/b/c2"},
			folder:           "/a/b/c2",
			absolutePath:     "/a/b/c2/Dockerfile",
			relativePath:     "Dockerfile",
		},
		{
			name:             "subfolder",
			uri:              "file:///a/b/c/d/Dockerfile",
			workspaceFolders: []string{"/a/b/c"},
			folder:           "/a/b/c",
			absolutePath:     "/a/b/c/d/Dockerfile",
			relativePath:     "d/Dockerfile",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			folder, absolutePath, relativePath := WorkspaceFolder(tc.uri, tc.workspaceFolders)
			require.Equal(t, tc.folder, folder)
			require.Equal(t, tc.absolutePath, absolutePath)
			require.Equal(t, tc.relativePath, relativePath)
		})
	}
}
