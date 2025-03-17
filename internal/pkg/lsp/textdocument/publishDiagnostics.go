package textdocument

import (
	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
)

type DiagnosticsCollector interface {
	CollectDiagnostics(source, workspaceFolder string, doc document.Document, text string) []protocol.Diagnostic
	SupportsLanguageIdentifier(languageIdentifier protocol.LanguageIdentifier) bool
}
