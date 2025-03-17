package document

import (
	"sync"

	"github.com/docker/docker-language-server/internal/bake/hcl/parser"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/hashicorp/hcl-lang/decoder"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"go.lsp.dev/uri"
)

type BakeHCLDocument interface {
	Document
	Decoder() *decoder.PathDecoder
	File() *hcl.File
}

func NewBakeHCLDocument(u uri.URI, version int32, input []byte) BakeHCLDocument {
	doc := &bakeHCLDocument{
		document: document{
			uri:        u,
			identifier: protocol.DockerBakeLanguage,
			version:    version,
			input:      input,
		},
	}
	doc.document.copyFn = doc.copy
	doc.document.parseFn = doc.parse
	doc.document.parseFn(true)
	return doc
}

type bakeHCLDocument struct {
	document
	mutex   sync.Mutex
	decoder *decoder.PathDecoder
	file    *hcl.File
}

func (d *bakeHCLDocument) parse(_ bool) bool {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	file, _ := hclsyntax.ParseConfig(d.document.input, string(d.document.uri), hcl.InitialPos)
	decoder := decoder.NewDecoder(&parser.PathReaderImpl{Filename: string(d.document.uri), File: file})
	pd, _ := decoder.Path(lang.Path{})
	pd.PrefillRequiredFields = true
	d.decoder = pd
	d.file = file
	return true
}

func (d *bakeHCLDocument) copy() Document {
	return NewBakeHCLDocument(d.uri, d.version, d.input)
}

func (d *dockerfileDocument) copy() Document {
	return NewDockerfileDocument(d.uri, d.version, d.input)
}

func (d *bakeHCLDocument) File() *hcl.File {
	return d.file
}

func (d *bakeHCLDocument) Decoder() *decoder.PathDecoder {
	return d.decoder
}
