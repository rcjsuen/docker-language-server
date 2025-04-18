package document

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"sync"

	"github.com/docker/buildx/bake"
	"github.com/docker/docker-language-server/internal/bake/hcl/parser"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/docker/docker-language-server/internal/types"
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
	DockerfileForTarget(block *hclsyntax.Block) (string, error)
}

type BakePrintOutput struct {
	Group  map[string]bake.Group  `json:"group,omitempty"`
	Target map[string]bake.Target `json:"target"`
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
	mutex           sync.Mutex
	decoder         *decoder.PathDecoder
	file            *hcl.File
	bakePrintOutput *BakePrintOutput
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
	d.extractBakeOutput()
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

func (d *bakeHCLDocument) extractBakeOutput() {
	body, ok := d.File().Body.(*hclsyntax.Body)
	if !ok {
		d.bakePrintOutput = nil
		return
	}

	targets := []string{}
	for _, b := range body.Blocks {
		if len(b.Labels) == 1 {
			if b.Type == "target" {
				targets = append(targets, b.Labels[0])
			}
		} else {
			// if the block's label count is not 1 Bake will not parse the file
			d.bakePrintOutput = nil
			return
		}
	}

	btargets, groups, err := bake.ReadTargets(
		context.Background(),
		[]bake.File{{Name: d.uri.Filename(), Data: d.Input()}},
		targets,
		nil,
		nil,
		nil,
	)
	if err != nil {
		d.bakePrintOutput = nil
		return
	}

	checkedGroups := map[string]bake.Group{}
	checkedTargets := map[string]bake.Target{}
	for target, value := range btargets {
		if value != nil {
			checkedTargets[target] = *value
		}
	}
	for group, value := range groups {
		if value != nil {
			checkedGroups[group] = *value
		}
	}
	d.bakePrintOutput = &BakePrintOutput{Group: checkedGroups, Target: checkedTargets}
}

func (d *bakeHCLDocument) DockerfileForTarget(block *hclsyntax.Block) (string, error) {
	if d.bakePrintOutput == nil || len(block.Labels) != 1 {
		return "", errors.New("cannot parse Bake file")
	}

	url, err := url.Parse(string(d.URI()))
	if err != nil {
		return "", fmt.Errorf("LSP client sent invalid URI: %v", string(d.URI()))
	}
	contextPath, err := types.AbsoluteFolder(url)
	if err != nil {
		return "", fmt.Errorf("LSP client sent invalid URI: %v", string(d.URI()))
	}

	if block, ok := d.bakePrintOutput.Target[block.Labels[0]]; ok {
		if block.DockerfileInline != nil {
			return "", nil
		} else if block.Context != nil {
			contextPath = *block.Context
			contextPath, err = types.AbsolutePath(url, contextPath)
			if err != nil {
				return "", nil
			}
		}

		if block.Dockerfile != nil {
			return filepath.Join(contextPath, *block.Dockerfile), nil
		}
		return filepath.Join(contextPath, "Dockerfile"), nil
	}
	return "", nil
}
