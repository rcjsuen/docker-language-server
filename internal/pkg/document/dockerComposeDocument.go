package document

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/docker/docker-language-server/internal/types"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"go.lsp.dev/uri"
)

type ComposeDocument interface {
	Document
	File() *ast.File
	ParsingError() error
	IncludedFiles() (map[string]*ast.File, bool)
}

type composeDocument struct {
	document
	mutex        sync.Mutex
	mgr          *Manager
	file         *ast.File
	parsingError error
}

func NewComposeDocument(mgr *Manager, u uri.URI, version int32, input []byte) ComposeDocument {
	doc := &composeDocument{
		document: document{
			uri:        u,
			identifier: protocol.DockerComposeLanguage,
			version:    version,
			input:      input,
		},
		mgr: mgr,
	}
	doc.document.copyFn = doc.copy
	doc.document.parseFn = doc.parse
	doc.document.parseFn(true)
	return doc
}

func (d *composeDocument) parse(_ bool) bool {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.file, d.parsingError = parser.ParseBytes(d.input, parser.ParseComments)
	return true
}

func (d *composeDocument) copy() Document {
	return NewComposeDocument(d.mgr, d.uri, d.version, d.input)
}

func (d *composeDocument) File() *ast.File {
	return d.file
}

func (d *composeDocument) ParsingError() error {
	return d.parsingError
}

func isPath(path string) bool {
	prefixes := []string{"git://", "http://", "https://", "oci://"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(path, prefix) {
			return false
		}
	}
	return true
}

func searchForIncludedFiles(searched []uri.URI, d *composeDocument) (map[string]*ast.File, bool) {
	u, err := url.Parse(string(d.uri))
	if err != nil {
		return nil, true
	}

	files := map[string]*ast.File{}
	for _, path := range d.includedPaths() {
		if isPath(path) {
			includedPath, err := types.AbsolutePath(u, path)
			if err == nil {
				uriString := fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(includedPath), "/"))
				pathURI := uri.URI(uriString)
				if slices.Contains(searched, pathURI) {
					return nil, false
				}
				doc, err := d.mgr.tryReading(context.Background(), pathURI, false)
				if err == nil {
					if c, ok := doc.(*composeDocument); ok && c.file != nil {
						searched = append(searched, pathURI)
						next, resolvable := searchForIncludedFiles(searched, c)
						if !resolvable {
							return nil, false
						}
						files[uriString] = c.file
						for u, f := range next {
							files[u] = f
						}
					}
				}
			}
		}
	}
	return files, true
}

func (d *composeDocument) IncludedFiles() (map[string]*ast.File, bool) {
	return searchForIncludedFiles([]uri.URI{d.uri}, d)
}

func (d *composeDocument) includedPaths() []string {
	if d.file == nil || len(d.file.Docs) == 0 {
		return nil
	}

	paths := []string{}
	for _, doc := range d.file.Docs {
		if node, ok := doc.Body.(*ast.MappingNode); ok {
			for _, topNode := range node.Values {
				if topNode.Key.GetToken().Value == "include" {
					for _, sequenceValue := range topNode.Value.(*ast.SequenceNode).Values {
						if stringArrayItem, ok := sequenceValue.(*ast.StringNode); ok {
							paths = append(paths, stringArrayItem.Value)
						} else if includeObject, ok := sequenceValue.(*ast.MappingNode); ok {
							for _, includeValues := range includeObject.Values {
								if includeValues.Key.GetToken().Value == "path" {
									if pathArrayItem, ok := includeValues.Value.(*ast.StringNode); ok {
										paths = append(paths, pathArrayItem.Value)
									} else if pathSequence, ok := includeValues.Value.(*ast.SequenceNode); ok {
										for _, sequenceValue := range pathSequence.Values {
											if s, ok := sequenceValue.(*ast.StringNode); ok {
												paths = append(paths, s.Value)
											}
										}
									}
									break
								}
							}
						}
					}
					break
				}
			}
		}
	}
	return paths
}
