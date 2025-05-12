package document

import (
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"go.lsp.dev/uri"
)

type Document interface {
	URI() uri.URI
	Copy() Document
	Input() []byte
	Version() int32
	Update(version int32, input []byte) bool
	Close()
	LanguageIdentifier() protocol.LanguageIdentifier
}

type NewDocumentFunc func(mgr *Manager, u uri.URI, identifier protocol.LanguageIdentifier, version int32, input []byte) Document

func NewDocument(mgr *Manager, u uri.URI, identifier protocol.LanguageIdentifier, version int32, input []byte) Document {
	if identifier == protocol.DockerBakeLanguage {
		return NewBakeHCLDocument(u, version, input)
	} else if identifier == protocol.DockerComposeLanguage {
		return NewComposeDocument(mgr, u, version, input)
	}
	return NewDockerfileDocument(u, version, input)
}

type document struct {
	uri        uri.URI
	identifier protocol.LanguageIdentifier
	version    int32
	// input is the file as it exists in the editor buffer.
	input   []byte
	parseFn func(force bool) bool
	copyFn  func() Document
}

var _ Document = &document{}

func (d *document) Update(version int32, input []byte) bool {
	d.version = version
	d.input = input
	return d.parseFn(true)
}

func (d *document) Version() int32 {
	return d.version
}

func (d *document) Input() []byte {
	return d.input
}

func (d *document) URI() uri.URI {
	return d.uri
}

func (d *document) LanguageIdentifier() protocol.LanguageIdentifier {
	return d.identifier
}

func (d *document) Close() {
}

// Copy creates a shallow copy of the Document.
//
// The Contents byte slice is returned as-is.
// A shallow copy of the Tree is made, as Tree-sitter trees are not thread-safe.
func (d *document) Copy() Document {
	return d.copyFn()
}
