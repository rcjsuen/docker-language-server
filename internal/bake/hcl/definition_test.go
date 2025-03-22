package hcl

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/stretchr/testify/require"
	"go.lsp.dev/uri"
)

func TestLocalDockerfileForNonWindows(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.SkipNow()
		return
	}

	u, err := url.Parse("file:///home/unix/docker-bake.hcl")
	require.NoError(t, err)
	path, err := LocalDockerfile(u)
	require.NoError(t, err)
	require.Equal(t, "/home/unix/Dockerfile", path)
}

func TestLocalDockerfileForWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.SkipNow()
		return
	}

	u, err := url.Parse("file:///c%3A/Users/windows/docker-bake.hcl")
	require.NoError(t, err)
	path, err := LocalDockerfile(u)
	require.NoError(t, err)
	require.Equal(t, "c:\\Users\\windows\\Dockerfile", path)
}

func TestDefinition(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(wd)))
	definitionTestFolderPath := filepath.Join(projectRoot, "testdata", "definition")

	dockerfilePath := filepath.Join(definitionTestFolderPath, "Dockerfile")
	bakeFilePath := filepath.Join(definitionTestFolderPath, "docker-bake.hcl")

	dockerfilePath = filepath.ToSlash(dockerfilePath)
	bakeFilePath = filepath.ToSlash(bakeFilePath)

	dockerfileURI := fmt.Sprintf("file:///%v", strings.TrimPrefix(dockerfilePath, "/"))
	bakeFileURI := fmt.Sprintf("file:///%v", strings.TrimPrefix(bakeFilePath, "/"))

	testCases := []struct {
		name      string
		content   string
		line      uint32
		character uint32
		links     any
	}{
		{
			name:      "reference valid stage with target attribute on the wrong character",
			content:   "target \"default\" {\ndockerfile = \"Dockerfile\"\ntarget = \"stage\" }",
			line:      2,
			character: 0, // point to the attribute's name instead of value
			links:     nil,
		},
		{
			name:      "reference valid stage with target attribute on the wrong line",
			content:   "target \"default\" {\ndockerfile = \"Dockerfile\"\ntarget = \"stage\" }",
			line:      1, // point to the dockerfile attribute instead of the target attribute
			character: 13,
			links:     nil,
		},
		{
			name:      "reference stage in a non-target attribute",
			content:   "target \"default\" {\n  network = \"stage\"\n}",
			line:      1,
			character: 17,
			links:     nil,
		},
		{
			name:      "reference stage in a variable block",
			content:   "variable \"var\" {\n  target = \"stage\"\n}",
			line:      1,
			character: 17,
			links:     nil,
		},
		{
			name:      "context attribute points to a declared variable",
			content:   "variable \"var\" {\n  default = \"stageName\"\n}\ntarget \"default\" {\n  context = var\n}",
			line:      4,
			character: 13,
			links: []protocol.Location{
				{
					URI: bakeFileURI,
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 10},
						End:   protocol.Position{Line: 0, Character: 13},
					},
				},
			},
		},
		{
			name:      "target attribute points to a stage defined by a declared variable",
			content:   "variable \"var\" {\n  default = \"stageName\"\n}\ntarget \"default\" {\n  target = var\n}",
			line:      4,
			character: 13,
			links: []protocol.Location{
				{
					URI: bakeFileURI,
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 10},
						End:   protocol.Position{Line: 0, Character: 13},
					},
				},
			},
		},
		{
			name:      "target attribute points to a stage defined by a declared variable without quotes",
			content:   "variable var {\n  default = \"stageName\"\n}\ntarget \"default\" {\n  target = var\n}",
			line:      4,
			character: 13,
			links: []protocol.Location{
				{
					URI: bakeFileURI,
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 9},
						End:   protocol.Position{Line: 0, Character: 12},
					},
				},
			},
		},
		{
			name:      "target attribute points to a stage defined by ${var} with quotes",
			content:   "variable var {\n  default = \"stageName\"\n}\ntarget \"default\" {\n  target = \"${var}\"\n}",
			line:      4,
			character: 15,
			links: []protocol.Location{
				{
					URI: bakeFileURI,
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 9},
						End:   protocol.Position{Line: 0, Character: 12},
					},
				},
			},
		},
		{
			name:      "target attribute points to a stage defined by an undeclared variable",
			content:   "target \"default\" {\n  target = undefinedVariable\n}",
			line:      1,
			character: 20,
			links:     nil,
		},
		{
			name:      "target attribute points to a top-level attribute",
			content:   "stageName = \"abc\"\ntarget \"default\" {\n  target = stageName\n}",
			line:      2,
			character: 16,
			links: []protocol.Location{
				{
					URI: bakeFileURI,
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 9},
					},
				},
			},
		},
		{
			name:      "inherits attribute points to a valid target",
			content:   "target \"source\" {}\ntarget \"default\" {\n  inherits = [ \"source\" ]\n}",
			line:      2,
			character: 20,
			links: []protocol.Location{
				{
					URI: bakeFileURI,
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 8},
						End:   protocol.Position{Line: 0, Character: 14},
					},
				},
			},
		},
		{
			name:      "group block's targets attribute points to a valid target",
			content:   "target \"t1\" {}\ngroup \"g1\" {\n  targets = [ \"t1\" ]\n}",
			line:      2,
			character: 16,
			links: []protocol.Location{
				{
					URI: bakeFileURI,
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 8},
						End:   protocol.Position{Line: 0, Character: 10},
					},
				},
			},
		},
		{
			name:      "inherits attribute points to an unquoted variable",
			content:   "variable \"var\" {}\ntarget \"default\" {\n  inherits = [ var ]\n}",
			line:      2,
			character: 17,
			links: []protocol.Location{
				{
					URI: bakeFileURI,
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 10},
						End:   protocol.Position{Line: 0, Character: 13},
					},
				},
			},
		},
		{
			name:      "inherits attribute points to a quoted variable",
			content:   "variable \"var\" {}\ntarget \"default\" {\n  inherits = [ \"${var}\" ]\n}",
			line:      2,
			character: 20,
			links: []protocol.Location{
				{
					URI: bakeFileURI,
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 10},
						End:   protocol.Position{Line: 0, Character: 13},
					},
				},
			},
		},
		{
			name:      "inherits attribute points to the second variable that is in quotes",
			content:   "variable \"var\" {}\nvariable \"var2\" {}\ntarget \"default\" {\n  inherits = [ var, \"${var2}\" ]\n}",
			line:      3,
			character: 24,
			links: []protocol.Location{
				{
					URI: bakeFileURI,
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 10},
						End:   protocol.Position{Line: 1, Character: 14},
					},
				},
			},
		},
		{
			name:      "inherits attribute points to the a quoted variable as the second item",
			content:   "variable \"var\" {}\ntarget \"default\" {\n  inherits = [ \"\", \"${var}\" ]\n}",
			line:      2,
			character: 24,
			links: []protocol.Location{
				{
					URI: bakeFileURI,
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 10},
						End:   protocol.Position{Line: 0, Character: 13},
					},
				},
			},
		},
		{
			name:      "inherits attribute pointing to a variable inside a quoted string should not work",
			content:   "variable \"source\" {}\ntarget \"default\" {\n  inherits = [ \"source\" ]\n}",
			line:      2,
			character: 20,
			links:     nil,
		},
		{
			name:      "entitlements attribute pointing to a variable",
			content:   "variable \"source\" {}\ntarget \"default\" {\n  entitlements = [ source ]\n}",
			line:      2,
			character: 22,
			links: []protocol.Location{
				{
					URI: bakeFileURI,
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 10},
						End:   protocol.Position{Line: 0, Character: 16},
					},
				},
			},
		},
		{
			name:      "inherits attribute pointing to a variable",
			content:   "variable \"source\" {}\ntarget \"default\" {\n  inherits = [ source ]\n}",
			line:      2,
			character: 18,
			links: []protocol.Location{
				{
					URI: bakeFileURI,
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 10},
						End:   protocol.Position{Line: 0, Character: 16},
					},
				},
			},
		},
		{
			name:      "formula referencing variable and top-level attribute with the location at the boolean check",
			content:   "default_network = \"none\"\nvariable \"networkType\" {\n  default = \"default\"\n}\ntarget \"default\" {\n  network = networkType == \"host\" ? networkType : default_network\n}",
			line:      5,
			character: 19,
			links: []protocol.Location{
				{
					URI: bakeFileURI,
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 10},
						End:   protocol.Position{Line: 1, Character: 21},
					},
				},
			},
		},
		{
			name:      "formula referencing variable and top-level attribute with the location at the true result",
			content:   "default_network = \"none\"\nvariable \"networkType\" {\n  default = \"default\"\n}\ntarget \"default\" {\n  network = networkType == \"host\" ? networkType : default_network\n}",
			line:      5,
			character: 43,
			links: []protocol.Location{
				{
					URI: bakeFileURI,
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 10},
						End:   protocol.Position{Line: 1, Character: 21},
					},
				},
			},
		},
		{
			name:      "formula referencing variable and top-level attribute with the location at the false result",
			content:   "default_network = \"none\"\nvariable \"networkType\" {\n  default = \"default\"\n}\ntarget \"default\" {\n  network = networkType == \"host\" ? networkType : default_network\n}",
			line:      5,
			character: 56,
			links: []protocol.Location{
				{
					URI: bakeFileURI,
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 15},
					},
				},
			},
		},
		{
			name:      "location in whitespace of a BinaryOpExpr",
			content:   "default_network = \"none\"\nnetworkType2 = \"none\"\ntarget \"default\" {\n  network = networkType  == networkType2 ? networkType : default_network\n}",
			line:      3,
			character: 24,
			links:     nil,
		},
		{
			name:      "location in whitespace of a ConditionalExpr",
			content:   "default_network = \"none\"\ntarget \"default\" {\n  network = networkType == \"host\" ? networkType :  default_network\n}",
			line:      2,
			character: 50,
			links:     nil,
		},
		{
			name:      "variable (referencing an attribute) inside an args attribute",
			content:   "var = \"value\"\ntarget \"default\" {\n  args = {\n    arg = var\n  }\n}",
			line:      3,
			character: 12,
			links: []protocol.Location{
				{
					URI: bakeFileURI,
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 3},
					},
				},
			},
		},
		{
			name:      "variable inside a function",
			content:   "variable \"TAG\" {}\ntarget \"default\" {\n  tags = [ notequal(\"\", TAG) ? \"image:${TAG}\" : \"image:latest\"\n}",
			line:      2,
			character: 26,
			links: []protocol.Location{
				{
					URI: bakeFileURI,
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 10},
						End:   protocol.Position{Line: 0, Character: 13},
					},
				},
			},
		},
		{
			name:      "${variable} inside a function",
			content:   "variable \"TAG\" {}\ntarget \"default\" {\n  tags = [ notequal(\"\", TAG) ? \"image:${TAG}\" : \"image:latest\"\n}",
			line:      2,
			character: 42,
			links: []protocol.Location{
				{
					URI: bakeFileURI,
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 10},
						End:   protocol.Position{Line: 0, Character: 13},
					},
				},
			},
		},
		{
			name:      "referenced function name",
			content:   "function \"tag\" {\n  params = [param]\n  result = [\"${param}\"]\n}\ntarget \"default\" {\n  tags = tag(\"v1\")\n}",
			line:      5,
			character: 10,
			links: []protocol.Location{
				{
					URI: bakeFileURI,
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 10},
						End:   protocol.Position{Line: 0, Character: 13},
					},
				},
			},
		},
		{
			name:      "referenced function name inside ${}",
			content:   "function \"tag\" {\n  params = [param]\n  result = [\"${param}\"]\n}\ntarget \"default\" {\n  tags = \"${tag(\"v1\")}\"\n}",
			line:      5,
			character: 15,
			links: []protocol.Location{
				{
					URI: bakeFileURI,
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 10},
						End:   protocol.Position{Line: 0, Character: 13},
					},
				},
			},
		},
		{
			name:      "attribute string value",
			content:   "a1 = \"value\"\n",
			line:      0,
			character: 9,
			links:     nil,
		},
		{
			name:      "attribute references another attribute",
			content:   "a1 = \"value\"\na2 = a1",
			line:      1,
			character: 6,
			links: []protocol.Location{
				{
					URI: bakeFileURI,
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 2},
					},
				},
			},
		},
		{
			name:      "attribute should point at itself",
			content:   "a1 = \"value\"\na2 = a1",
			line:      1,
			character: 1,
			links: []protocol.Location{
				{
					URI: bakeFileURI,
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 0},
						End:   protocol.Position{Line: 1, Character: 2},
					},
				},
			},
		},
		{
			name:      "variable referenced in for loop conditional",
			content:   "variable num { default = 3 }\nvariable varList { default = [\"tag\"] }\ntarget default {\n  tags = [for var in varList : upper(var) if num > 2]\n}",
			line:      3,
			character: 46,
			links: []protocol.Location{
				{
					URI: bakeFileURI,
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 9},
						End:   protocol.Position{Line: 0, Character: 12},
					},
				},
			},
		},
		{
			name:      "variable inside a for loop",
			content:   "variable varList { default = [\"tag\"] }\ntarget default {\n  tags = [for var in varList : upper(var)]\n}",
			line:      2,
			character: 24,
			links: []protocol.Location{
				{
					URI: bakeFileURI,
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 9},
						End:   protocol.Position{Line: 0, Character: 16},
					},
				},
			},
		},
		{
			name:      "args key references Dockerfile ARG variable (unquoted key, no default value set)",
			content:   "target default {\n  args = {\n    var = \"value\"\n  }\n}",
			line:      2,
			character: 6,
			links: []protocol.Location{
				{
					URI: dockerfileURI,
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 0},
						End:   protocol.Position{Line: 1, Character: 7},
					},
				},
			},
		},
		{
			name:      "args key references Dockerfile ARG variable (unquoted key, default value set)",
			content:   "target default {\n  args = {\n    defined = \"value\"\n  }\n}",
			line:      2,
			character: 8,
			links: []protocol.Location{
				{
					URI: dockerfileURI,
					Range: protocol.Range{
						Start: protocol.Position{Line: 2, Character: 0},
						End:   protocol.Position{Line: 2, Character: 19},
					},
				},
			},
		},
		{
			name:      "args key references Dockerfile ARG variable (quoted key, no default value set)",
			content:   "target default {\n  args = {\n    \"var\" = \"value\"\n  }\n}",
			line:      2,
			character: 7,
			links: []protocol.Location{
				{
					URI: dockerfileURI,
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 0},
						End:   protocol.Position{Line: 1, Character: 7},
					},
				},
			},
		},
		{
			name:      "group block with an invalid inherits attribute should not return a result",
			content:   "target t1 {}\ngroup g1 { inherits = [\"t1\"] }",
			line:      1,
			character: 25,
			links:     nil,
		},
		{
			name:      "variable block with an invalid inherits attribute should not return a result",
			content:   "target t1 {}\nvariable v1 { inherits = [\"t1\"] }",
			line:      1,
			character: 28,
			links:     nil,
		},
		{
			name:      "args key should not reference in a group block",
			content:   "group g1 {\n  args = {\n    var = \"value\"\n  }\n}",
			line:      2,
			character: 6,
			links:     nil,
		},
		{
			name:      "args key should not reference in a variable block",
			content:   "variable var {\n  args = {\n    var = \"value\"\n  }\n}",
			line:      2,
			character: 6,
			links:     nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			manager := document.NewDocumentManager()
			doc := document.NewBakeHCLDocument(uri.URI(bakeFileURI), 1, []byte(tc.content))
			links, err := Definition(context.Background(), true, manager, uri.URI(bakeFileURI), doc, protocol.Position{Line: tc.line, Character: tc.character})
			require.NoError(t, err)
			require.Equal(t, tc.links, links)
		})
	}
}

func TestDefinitionVariedResults(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(wd)))
	definitionTestFolderPath := filepath.Join(projectRoot, "testdata", "definition")

	dockerfilePath := filepath.Join(definitionTestFolderPath, "Dockerfile")
	bakeFilePath := filepath.Join(definitionTestFolderPath, "docker-bake.hcl")

	dockerfilePath = filepath.ToSlash(dockerfilePath)
	bakeFilePath = filepath.ToSlash(bakeFilePath)

	dockerfileURI := fmt.Sprintf("file:///%v", strings.TrimPrefix(dockerfilePath, "/"))
	bakeFileURI := fmt.Sprintf("file:///%v", strings.TrimPrefix(bakeFilePath, "/"))

	testCases := []struct {
		name      string
		content   string
		line      uint32
		character uint32
		locations any
		links     any
	}{
		{
			name:      "reference valid stage (target block, target attribute)",
			content:   "target \"default\" { target = \"stage\" }",
			line:      0,
			character: 32,
			locations: []protocol.Location{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 21},
					},
					URI: dockerfileURI,
				},
			},
			links: []protocol.LocationLink{
				{
					OriginSelectionRange: &protocol.Range{
						Start: protocol.Position{Line: 0, Character: 29},
						End:   protocol.Position{Line: 0, Character: 34},
					},
					TargetURI: dockerfileURI,
					TargetRange: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 21},
					},
					TargetSelectionRange: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 21},
					},
				},
			},
		},
		{
			name:      "hyphenated stage is highlighted completely (target block, target attribute)",
			content:   "target \"default\" { target = \"hyphenated-stage\" }",
			line:      0,
			character: 33,
			locations: []protocol.Location{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 3, Character: 0},
						End:   protocol.Position{Line: 3, Character: 32},
					},
					URI: dockerfileURI,
				},
			},
			links: []protocol.LocationLink{
				{
					OriginSelectionRange: &protocol.Range{
						Start: protocol.Position{Line: 0, Character: 29},
						End:   protocol.Position{Line: 0, Character: 45},
					},
					TargetURI: dockerfileURI,
					TargetRange: protocol.Range{
						Start: protocol.Position{Line: 3, Character: 0},
						End:   protocol.Position{Line: 3, Character: 32},
					},
					TargetSelectionRange: protocol.Range{
						Start: protocol.Position{Line: 3, Character: 0},
						End:   protocol.Position{Line: 3, Character: 32},
					},
				},
			},
		},
		{
			name:      "reference valid stage (target block, no-cache-filter attribute)",
			content:   "target \"default\" { no-cache-filter = [\"stage\"] }",
			line:      0,
			character: 42,
			locations: []protocol.Location{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 21},
					},
					URI: dockerfileURI,
				},
			},
			links: []protocol.LocationLink{
				{
					OriginSelectionRange: &protocol.Range{
						Start: protocol.Position{Line: 0, Character: 39},
						End:   protocol.Position{Line: 0, Character: 44},
					},
					TargetURI: dockerfileURI,
					TargetRange: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 21},
					},
					TargetSelectionRange: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 21},
					},
				},
			},
		},
		{
			name:      "reference valid stage with target attribute on the right position",
			content:   "target \"default\" {\ndockerfile = \"Dockerfile\"\ntarget = \"stage\" }",
			line:      2,
			character: 13,
			locations: []protocol.Location{
				{
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 21},
					},
					URI: dockerfileURI,
				},
			},
			links: []protocol.LocationLink{
				{
					OriginSelectionRange: &protocol.Range{
						Start: protocol.Position{Line: 2, Character: 10},
						End:   protocol.Position{Line: 2, Character: 15},
					},
					TargetURI: dockerfileURI,
					TargetRange: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 21},
					},
					TargetSelectionRange: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 21},
					},
				},
			},
		},
		{
			name:      "no-cache-filter attribute should not work in a group block",
			content:   "group \"g1\" { no-cache-filter = [\"stage\"] }",
			line:      0,
			character: 39,
			locations: nil,
			links:     nil,
		},
		{
			name:      "no-cache-filter attribute should not work in a variable block",
			content:   "variable \"var\" {\ndockerfile = \"Dockerfile\"\ntarget = \"stage\" }",
			line:      2,
			character: 13,
			locations: nil,
			links:     nil,
		},
		{
			name:      "target attribute should not work in a group block",
			content:   "group \"g1\" { target = \"stage\" }",
			line:      0,
			character: 25,
			locations: nil,
			links:     nil,
		},
		{
			name:      "target attribute should not work in a variable block",
			content:   "variable \"var\" { target = \"stage\" }",
			line:      0,
			character: 30,
			locations: nil,
			links:     nil,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v (Location)", tc.name), func(t *testing.T) {
			manager := document.NewDocumentManager()
			doc := document.NewBakeHCLDocument(uri.URI(bakeFileURI), 1, []byte(tc.content))
			locations, err := Definition(context.Background(), false, manager, uri.URI(bakeFileURI), doc, protocol.Position{Line: tc.line, Character: tc.character})
			require.NoError(t, err)
			require.Equal(t, tc.locations, locations)
		})

		t.Run(fmt.Sprintf("%v (LocationLink)", tc.name), func(t *testing.T) {
			manager := document.NewDocumentManager()
			doc := document.NewBakeHCLDocument(uri.URI(bakeFileURI), 1, []byte(tc.content))
			links, err := Definition(context.Background(), true, manager, uri.URI(bakeFileURI), doc, protocol.Position{Line: tc.line, Character: tc.character})
			require.NoError(t, err)
			require.Equal(t, tc.links, links)
		})
	}
}
