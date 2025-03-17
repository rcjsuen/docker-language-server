package document

import (
	"sync"

	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"go.lsp.dev/uri"
	"gopkg.in/yaml.v3"
)

type ComposeDocument interface {
	Document
	RootNode() yaml.Node
}

type composeDocument struct {
	document
	mutex    sync.Mutex
	rootNode yaml.Node
}

func NewComposeDocument(u uri.URI, version int32, input []byte) ComposeDocument {
	doc := &composeDocument{
		document: document{
			uri:        u,
			identifier: protocol.DockerComposeLanguage,
			version:    version,
			input:      input,
		},
	}
	doc.document.copyFn = doc.copy
	doc.document.parseFn = doc.parse
	doc.document.parseFn(true)
	return doc
}

func (d *composeDocument) parse(_ bool) bool {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	_ = yaml.Unmarshal([]byte(d.input), &d.rootNode)
	return true
}

func (d *composeDocument) copy() Document {
	return NewComposeDocument(d.uri, d.version, d.input)
}

func (d *composeDocument) RootNode() yaml.Node {
	return d.rootNode
}
