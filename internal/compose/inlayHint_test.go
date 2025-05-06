package compose

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/docker/docker-language-server/internal/types"
	"github.com/stretchr/testify/require"
	"go.lsp.dev/uri"
)

func TestInlayHint(t *testing.T) {
	composeFileURI := fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(filepath.Join(os.TempDir(), "compose.yaml")), "/"))

	testCases := []struct {
		name       string
		content    string
		inlayHints []protocol.InlayHint
	}{
		{
			name: "extends is ignored",
			content: `
services:
  web:
    attach: true
  web2:
    extends: web
  web3:
    extends: web2`,
			inlayHints: []protocol.InlayHint{},
		},
		{
			name: "single line attribute has a hint",
			content: `
services:
  web:
    attach: true
  web2:
    extends: web
    attach: false`,
			inlayHints: []protocol.InlayHint{
				{
					Label:       "(parent value: true)",
					PaddingLeft: types.CreateBoolPointer(true),
					Position:    protocol.Position{Line: 6, Character: 17},
				},
			},
		},
		{
			name: "attribute recurses upwards",
			content: `
services:
  web:
    attach: true
  web2:
    extends: web
  web3:
    extends: web
    attach: false`,
			inlayHints: []protocol.InlayHint{
				{
					Label:       "(parent value: true)",
					PaddingLeft: types.CreateBoolPointer(true),
					Position:    protocol.Position{Line: 8, Character: 17},
				},
			},
		},
		{
			name: "extends as an object but without a file attribute",
			content: `
services:
  web:
    attach: true
  web2:
    extends:
      service: web
    attach: false`,
			inlayHints: []protocol.InlayHint{
				{
					Label:       "(parent value: true)",
					PaddingLeft: types.CreateBoolPointer(true),
					Position:    protocol.Position{Line: 7, Character: 17},
				},
			},
		},
		{
			name: "extends as an object pointing to a locally named service but points to a bad file",
			content: `
services:
  web:
    attach: true
  web2:
    extends:
      service: web
      file: non-existent.yaml
    attach: false`,
			inlayHints: []protocol.InlayHint{},
		},
		{
			name: "quoted string value has the correct position",
			content: `
services:
  web:
    hostname: "hostname1"
  web2:
    hostname: "hostname2"
    extends: web`,
			inlayHints: []protocol.InlayHint{
				{
					Label:       "(parent value: hostname1)",
					PaddingLeft: types.CreateBoolPointer(true),
					Position:    protocol.Position{Line: 5, Character: 25},
				},
			},
		},
		{
			name: "sub-attributes unsupported",
			content: `
services:
  web:
    build:
      context: c1
  web2:
    build:
      context: c2
    extends: web`,
			inlayHints: []protocol.InlayHint{},
		},
	}

	for _, tc := range testCases {
		u := uri.URI(composeFileURI)
		t.Run(tc.name, func(t *testing.T) {
			doc := document.NewComposeDocument(u, 1, []byte(tc.content))
			inlayHints, err := InlayHint(doc, protocol.Range{})
			require.NoError(t, err)
			require.Equal(t, tc.inlayHints, inlayHints)
		})
	}
}
