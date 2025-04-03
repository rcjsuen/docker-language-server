package server

import (
	"context"
	"hash/fnv"
	"os"

	"github.com/docker/docker-language-server/internal/configuration"
	"github.com/docker/docker-language-server/internal/telemetry"
	"github.com/docker/docker-language-server/internal/tliron/glsp"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/docker/docker-language-server/internal/types"
	"go.lsp.dev/uri"
)

func (s *Server) TextDocumentDidOpen(ctx *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
	_, _ = s.docs.Write(ctx.Context, uri.URI(params.TextDocument.URI), params.TextDocument.LanguageID, params.TextDocument.Version, []byte(params.TextDocument.Text))
	go func() {
		defer s.handlePanic("TextDocumentDidOpen")

		s.FetchConfigurations([]protocol.DocumentUri{params.TextDocument.URI})
		s.computeDiagnostics(ctx.Context, params.TextDocument.URI)
	}()
	return nil
}

func (s *Server) TextDocumentDidChange(ctx *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
	if len(params.ContentChanges) == 0 {
		return nil
	}

	if changeEvent, ok := params.ContentChanges[0].(protocol.TextDocumentContentChangeEvent); ok {
		changed, _ := s.docs.Overwrite(ctx.Context, uri.URI(params.TextDocument.URI), params.TextDocument.Version, []byte(changeEvent.Text))
		if changed {
			s.computeDiagnostics(ctx.Context, params.TextDocument.URI)
		}
	} else if changeEventWhole, ok := params.ContentChanges[0].(protocol.TextDocumentContentChangeEventWhole); ok {
		changed, _ := s.docs.Overwrite(ctx.Context, uri.URI(params.TextDocument.URI), params.TextDocument.Version, []byte(changeEventWhole.Text))
		if changed {
			s.computeDiagnostics(ctx.Context, params.TextDocument.URI)
		}
	}
	return nil
}

func (s *Server) TextDocumentDidClose(ctx *glsp.Context, params *protocol.DidCloseTextDocumentParams) error {
	s.docs.Remove(uri.URI(params.TextDocument.URI))

	go func() {
		defer s.handlePanic("TextDocumentDidClose")
		configuration.Remove(params.TextDocument.URI)
		// clear out all existing diagnostics when the editor has been closed
		s.client.PublishDiagnostics(context.Background(), protocol.PublishDiagnosticsParams{
			URI:         params.TextDocument.URI,
			Diagnostics: []protocol.Diagnostic{},
		})
	}()
	return nil
}

func (s *Server) computeDiagnostics(ctx context.Context, documentURI protocol.DocumentUri) {
	doc := s.docs.Get(ctx, uri.URI(documentURI))
	folder, absolutePath, relativePath := types.WorkspaceFolder(documentURI, s.workspaceFolders)
	if folder == "" {
		s.recordAnalysis(doc.LanguageIdentifier(), "unversioned", absolutePath)
	} else {
		remote := s.gitRemotes[folder]
		if remote == "" {
			s.recordAnalysis(doc.LanguageIdentifier(), "unversioned", absolutePath)
		} else {
			s.recordAnalysis(doc.LanguageIdentifier(), remote, relativePath)
		}
	}

	s.docs.Queue(ctx, uri.URI(documentURI), func() {
		defer s.handlePanic("computeDiagnostics")
		doc := s.docs.Get(context.Background(), uri.URI(documentURI))

		folder = types.StripLeadingSlash(folder)
		diagnostics := []protocol.Diagnostic{}
		for _, collector := range s.diagnosticsCollectors {
			if collector.SupportsLanguageIdentifier(doc.LanguageIdentifier()) {
				if folder == "" {
					diagnostics = append(diagnostics, collector.CollectDiagnostics("docker-language-server", os.TempDir(), doc, string(doc.Input()))...)
				} else {
					diagnostics = append(diagnostics, collector.CollectDiagnostics("docker-language-server", folder, doc, string(doc.Input()))...)
				}
			}
		}

		version := doc.Version()
		s.client.PublishDiagnostics(context.Background(), protocol.PublishDiagnosticsParams{
			URI:         documentURI,
			Diagnostics: diagnostics,
			Version:     &version,
		})
	})
}

// recordAnalysis queues a telemetry event to record that the given path
// under the specified Git remote has been analyzed. gitRemote and path
// will be hashed before it is sent to the telemetry backend.
func (s *Server) recordAnalysis(languageIdentifier protocol.LanguageIdentifier, gitRemote, path string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if _, ok := s.analyzedFiles[gitRemote]; !ok {
		s.analyzedFiles[gitRemote] = make(map[protocol.DocumentUri]bool)
	}

	if _, ok := s.analyzedFiles[gitRemote][path]; !ok {
		s.analyzedFiles[gitRemote][path] = true

		go func() {
			defer s.handlePanic("recordAnalysis")

			hasher := fnv.New32a()
			hasher.Write([]byte(gitRemote))
			remote := hasher.Sum32()

			hasher.Reset()
			hasher.Write([]byte(path))
			document := hasher.Sum32()

			s.Enqueue(telemetry.EventServerUserAction, map[string]any{
				"action":              telemetry.ServerUserActionTypeFileAnalyzed,
				"language_identifier": string(languageIdentifier),
				"git_remote":          remote,
				"document":            document,
			})
		}()
	}
}
