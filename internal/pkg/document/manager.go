package document

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/bep/debounce"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"go.lsp.dev/uri"
)

type ManagerOpt func(manager *Manager)
type ReadDocumentFunc func(uri.URI) ([]byte, error)
type DocumentMap map[uri.URI]Document

// Manager provides simplified file read/write operations for the LSP server.
type Manager struct {
	mu                    sync.Mutex
	docs                  DocumentMap
	diagnosticsProcessing map[uri.URI]func(func())
	newDocFunc            NewDocumentFunc
	readDocFunc           ReadDocumentFunc
}

func NewDocumentManager(opts ...ManagerOpt) *Manager {
	m := Manager{
		docs:                  make(DocumentMap),
		diagnosticsProcessing: make(map[uri.URI]func(func())),
		newDocFunc:            NewDocument,
		readDocFunc:           ReadDocument,
	}

	for _, opt := range opts {
		opt(&m)
	}

	return &m
}

func WithReadDocumentFunc(readDocFunc ReadDocumentFunc) ManagerOpt {
	return func(manager *Manager) {
		manager.readDocFunc = readDocFunc
	}
}

// Read the document from the given URI and return its contents. This default
// implementation of a ReadDocumentFunc only handles file: URIs and returns an
// error otherwise.
func ReadDocument(u uri.URI) (contents []byte, err error) {
	fn, err := filename(u)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(fn)
}

func filename(u uri.URI) (fn string, err error) {
	defer func() {
		// recover from non-file URI in uri.Filename()
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	return u.Filename(), err
}

func (m *Manager) Version(ctx context.Context, u uri.URI) (int32, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if doc, found := m.docs[u]; found {
		return doc.Version(), nil
	}
	return -1, errors.New("document not managed")
}

// Read returns the contents of the file for the given URI.
//
// If no file exists at the path or the URI is of an invalid type, an error is
// returned.
func (m *Manager) Read(ctx context.Context, u uri.URI) (doc Document, err error) {
	m.mu.Lock()
	defer func() {
		if err == nil {
			// Always return a copy of the document
			doc = doc.Copy()
		}
		m.mu.Unlock()
	}()

	// TODO(siegs): check staleness for files read from disk?
	var found bool
	if doc, found = m.docs[u]; !found {
		_, err = m.readAndParse(ctx, u)
		doc = m.docs[u]
	}

	if os.IsNotExist(err) {
		err = os.ErrNotExist
	}

	return doc, err
}

// Queue enqueues the given function as something that should be run in
// the near future. The URI will be used as a key so if another function
// had previously enqueued for this function and it had not yet been run
// that the previously enqueued function will be discarded and replaced
// with the provided function.
func (m *Manager) Queue(ctx context.Context, u uri.URI, fn func()) {
	m.mu.Lock()
	defer m.mu.Unlock()
	queue := m.diagnosticsProcessing[u]
	queue(fn)
}

func (m *Manager) Get(ctx context.Context, u uri.URI) Document {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.docs[u]
}

// Write creates or replaces the contents of the file for the given URI.
// True will be returned if the document's syntax tree has changed.
func (m *Manager) Write(ctx context.Context, u uri.URI, identifier protocol.LanguageIdentifier, version int32, input []byte) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.write(ctx, u, identifier, version, input)
}

func (m *Manager) write(ctx context.Context, u uri.URI, identifier protocol.LanguageIdentifier, version int32, input []byte) (bool, error) {
	changed, err := m.parse(ctx, u, identifier, version, input)
	if err != nil {
		return false, fmt.Errorf("could not parse file %q: %v", u, err)
	}
	return changed, err
}

func (m *Manager) Overwrite(ctx context.Context, u uri.URI, version int32, input []byte) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if document, found := m.docs[u]; !found {
		changed, err := m.parse(ctx, u, document.LanguageIdentifier(), version, input)
		if err != nil {
			return false, fmt.Errorf("could not parse file %q: %v", u, err)
		}
		return changed, nil
	}

	identifier := protocol.DockerfileLanguage
	if strings.HasSuffix(string(u), "hcl") {
		identifier = protocol.DockerBakeLanguage
	}
	return m.write(ctx, u, identifier, version, input)
}

func (m *Manager) Remove(u uri.URI) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.removeAndCleanup(u)
}

func (m *Manager) Keys() []uri.URI {
	m.mu.Lock()
	defer m.mu.Unlock()
	keys := make([]uri.URI, 0, len(m.docs))
	for k := range m.docs {
		keys = append(keys, k)
	}
	return keys
}

func (m *Manager) readAndParse(ctx context.Context, u uri.URI) (bool, error) {
	identifier := protocol.DockerfileLanguage
	if strings.HasSuffix(string(u), "hcl") {
		identifier = protocol.DockerBakeLanguage
	}

	if _, found := m.docs[u]; !found {
		contents, err := m.readDocFunc(u)
		if err != nil {
			return false, err
		}
		return m.parse(ctx, u, identifier, 1, contents)
	}
	return m.parse(ctx, u, identifier, 1, nil)
}

func (m *Manager) parse(_ context.Context, uri uri.URI, identifier protocol.LanguageIdentifier, version int32, input []byte) (bool, error) {
	doc, loaded := m.docs[uri]
	changed := true
	if !loaded {
		doc = m.newDocFunc(uri, identifier, version, input)
		m.docs[uri] = doc
		m.diagnosticsProcessing[uri] = debounce.New(time.Millisecond * 50)
	} else {
		changed = doc.Update(version, input)
	}

	return changed, nil
}

// removeAndCleanup removes a Document and frees associated resources.
func (m *Manager) removeAndCleanup(uri uri.URI) {
	if existing, ok := m.docs[uri]; ok {
		existing.Close()
	}
	if fn, ok := m.diagnosticsProcessing[uri]; ok {
		// resets the debounced function to avoid parsing a document
		// that has already been closed
		fn(func() {})
	}
	delete(m.diagnosticsProcessing, uri)
	delete(m.docs, uri)
}
