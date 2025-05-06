package compose

import (
	"fmt"
	"math"
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

func TestCollectDiagnostics(t *testing.T) {
	testCases := []struct {
		name        string
		content     string
		diagnostics []protocol.Diagnostic
	}{
		{
			name: "tab is flagged as an error",
			content: `
service:
	abc:`,
			diagnostics: []protocol.Diagnostic{
				{
					Message:  "found character '\t' that cannot start any token",
					Source:   types.CreateStringPointer("docker-language-server"),
					Severity: types.CreateDiagnosticSeverityPointer(protocol.DiagnosticSeverityError),
					Range: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 0},
						End:   protocol.Position{Line: 2, Character: math.MaxUint32},
					},
				},
			},
		},
	}

	composeFileURI := uri.URI(fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(filepath.Join(os.TempDir(), "compose.yaml")), "/")))
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			collector := NewComposeDiagnosticsCollector()
			doc := document.NewComposeDocument(composeFileURI, 1, []byte(tc.content))
			diagnostics := collector.CollectDiagnostics("docker-language-server", "", doc, "")
			require.Equal(t, tc.diagnostics, diagnostics)
		})
	}
}
