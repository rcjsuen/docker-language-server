package scout

import (
	"strings"

	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/docker/docker-language-server/internal/types"
)

type ImageRequest struct {
	Image        string `json:"image"`
	Organization string `json:"organization"`
}

type Image struct {
	Short      string `json:"short"`
	Registry   string `json:"registry"`
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
	Digest     string `json:"digest"`
}

type Link struct {
	Href  string `json:"href"`
	Title string `json:"title"`
}

type Diagnostic struct {
	Kind     string `json:"kind"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
	Link     Link   `json:"link"`
}

type Edit struct {
	Title      string `json:"title"`
	Edit       string `json:"edit"`
	Diagnostic string `json:"diagnostic"`
}

type Description struct {
	Plaintext string `json:"plaintext"`
	Markdown  string `json:"markdown"`
}

type Info struct {
	Kind        string      `json:"kind"`
	Description Description `json:"description"`
}

type ImageResponse struct {
	Image       Image        `json:"image"`
	Diagnostics []Diagnostic `json:"diagnostics"`
	Edits       []Edit       `json:"edits"`
	Infos       []Info       `json:"infos"`
}

type ScoutImageKey struct {
	Image string
}

func (k *ScoutImageKey) CacheKey() string {
	return k.Image
}

func ConvertDiagnostic(diagnostic Diagnostic, words []string, source string, rng protocol.Range, edits []Edit) protocol.Diagnostic {
	lspDiagnostic := protocol.Diagnostic{}
	lspDiagnostic.Code = &protocol.IntegerOrString{Value: diagnostic.Kind}
	lspDiagnostic.Message = diagnostic.Message
	lspDiagnostic.Source = types.CreateStringPointer(source)
	if diagnostic.Link.Href != "" {
		lspDiagnostic.CodeDescription = &protocol.CodeDescription{
			HRef: diagnostic.Link.Href,
		}
	}
	lspDiagnostic.Range = rng

	switch diagnostic.Severity {
	case "error":
		lspDiagnostic.Severity = types.CreateDiagnosticSeverityPointer(protocol.DiagnosticSeverityError)
	case "info":
		lspDiagnostic.Severity = types.CreateDiagnosticSeverityPointer(protocol.DiagnosticSeverityInformation)
	case "warn":
		lspDiagnostic.Severity = types.CreateDiagnosticSeverityPointer(protocol.DiagnosticSeverityWarning)
	case "hint":
		lspDiagnostic.Severity = types.CreateDiagnosticSeverityPointer(protocol.DiagnosticSeverityHint)
	}

	includedEdits := []types.NamedEdit{}
	for _, edit := range edits {
		if lspDiagnostic.Code.Value == edit.Diagnostic {
			words[1] = edit.Edit
			edit.Edit = strings.Join(words, " ")
			includedEdits = append(includedEdits, types.NamedEdit{
				Title: edit.Title,
				Edit:  edit.Edit,
			})
		}
	}

	if len(includedEdits) > 0 {
		lspDiagnostic.Data = includedEdits
	}
	return lspDiagnostic
}
