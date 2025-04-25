package compose

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/docker/docker-language-server/internal/pkg/document"
	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
	"github.com/docker/docker-language-server/internal/types"
	"github.com/stretchr/testify/require"
	"go.lsp.dev/uri"
)

var serviceProperties = []protocol.CompletionItem{
	{Label: "annotations", Detail: types.CreateStringPointer("array or object")},
	{Label: "attach", Detail: types.CreateStringPointer("boolean or string")},
	{Label: "blkio_config", Detail: types.CreateStringPointer("object")},
	{Label: "build", Detail: types.CreateStringPointer("object or string")},
	{Label: "cap_add", Detail: types.CreateStringPointer("array")},
	{Label: "cap_drop", Detail: types.CreateStringPointer("array")},
	{Label: "cgroup", Detail: types.CreateStringPointer("string")},
	{Label: "cgroup_parent", Detail: types.CreateStringPointer("string")},
	{Label: "command", Detail: types.CreateStringPointer("array or null or string")},
	{Label: "configs", Detail: types.CreateStringPointer("array")},
	{Label: "container_name", Detail: types.CreateStringPointer("string")},
	{Label: "cpu_count", Detail: types.CreateStringPointer("integer or string")},
	{Label: "cpu_percent", Detail: types.CreateStringPointer("integer or string")},
	{Label: "cpu_period", Detail: types.CreateStringPointer("number or string")},
	{Label: "cpu_quota", Detail: types.CreateStringPointer("number or string")},
	{Label: "cpu_rt_period", Detail: types.CreateStringPointer("number or string")},
	{Label: "cpu_rt_runtime", Detail: types.CreateStringPointer("number or string")},
	{Label: "cpu_shares", Detail: types.CreateStringPointer("number or string")},
	{Label: "cpus", Detail: types.CreateStringPointer("number or string")},
	{Label: "cpuset", Detail: types.CreateStringPointer("string")},
	{Label: "credential_spec", Detail: types.CreateStringPointer("object")},
	{Label: "depends_on", Detail: types.CreateStringPointer("array or object")},
	{Label: "deploy", Detail: types.CreateStringPointer("null or object")},
	{Label: "develop", Detail: types.CreateStringPointer("null or object")},
	{Label: "device_cgroup_rules", Detail: types.CreateStringPointer("array")},
	{Label: "devices", Detail: types.CreateStringPointer("array")},
	{Label: "dns", Detail: types.CreateStringPointer("array or string")},
	{Label: "dns_opt", Detail: types.CreateStringPointer("array")},
	{Label: "dns_search", Detail: types.CreateStringPointer("array or string")},
	{Label: "domainname", Detail: types.CreateStringPointer("string")},
	{Label: "entrypoint", Detail: types.CreateStringPointer("array or null or string")},
	{Label: "env_file", Detail: types.CreateStringPointer("array or string")},
	{Label: "environment", Detail: types.CreateStringPointer("array or object")},
	{Label: "expose", Detail: types.CreateStringPointer("array")},
	{Label: "extends", Detail: types.CreateStringPointer("object or string")},
	{Label: "external_links", Detail: types.CreateStringPointer("array")},
	{Label: "extra_hosts", Detail: types.CreateStringPointer("array or object")},
	{Label: "gpus", Detail: types.CreateStringPointer("array or string")},
	{Label: "group_add", Detail: types.CreateStringPointer("array")},
	{Label: "healthcheck", Detail: types.CreateStringPointer("object")},
	{Label: "hostname", Detail: types.CreateStringPointer("string")},
	{Label: "image", Detail: types.CreateStringPointer("string")},
	{Label: "init", Detail: types.CreateStringPointer("boolean or string")},
	{Label: "ipc", Detail: types.CreateStringPointer("string")},
	{Label: "isolation", Detail: types.CreateStringPointer("string")},
	{Label: "label_file", Detail: types.CreateStringPointer("array or string")},
	{Label: "labels", Detail: types.CreateStringPointer("array or object")},
	{Label: "links", Detail: types.CreateStringPointer("array")},
	{Label: "logging", Detail: types.CreateStringPointer("object")},
	{Label: "mac_address", Detail: types.CreateStringPointer("string")},
	{Label: "mem_limit", Detail: types.CreateStringPointer("number or string")},
	{Label: "mem_reservation", Detail: types.CreateStringPointer("integer or string")},
	{Label: "mem_swappiness", Detail: types.CreateStringPointer("integer or string")},
	{Label: "memswap_limit", Detail: types.CreateStringPointer("number or string")},
	{Label: "network_mode", Detail: types.CreateStringPointer("string")},
	{Label: "networks", Detail: types.CreateStringPointer("array or object")},
	{Label: "oom_kill_disable", Detail: types.CreateStringPointer("boolean or string")},
	{Label: "oom_score_adj", Detail: types.CreateStringPointer("integer or string")},
	{Label: "pid", Detail: types.CreateStringPointer("null or string")},
	{Label: "pids_limit", Detail: types.CreateStringPointer("number or string")},
	{Label: "platform", Detail: types.CreateStringPointer("string")},
	{Label: "ports", Detail: types.CreateStringPointer("array")},
	{Label: "post_start", Detail: types.CreateStringPointer("array")},
	{Label: "pre_stop", Detail: types.CreateStringPointer("array")},
	{Label: "privileged", Detail: types.CreateStringPointer("boolean or string")},
	{Label: "profiles", Detail: types.CreateStringPointer("array")},
	{Label: "provider", Detail: types.CreateStringPointer("object")},
	{Label: "pull_policy", Detail: types.CreateStringPointer("string")},
	{Label: "pull_refresh_after", Detail: types.CreateStringPointer("string")},
	{Label: "read_only", Detail: types.CreateStringPointer("boolean or string")},
	{Label: "restart", Detail: types.CreateStringPointer("string")},
	{Label: "runtime", Detail: types.CreateStringPointer("string")},
	{Label: "scale", Detail: types.CreateStringPointer("integer or string")},
	{Label: "secrets", Detail: types.CreateStringPointer("array")},
	{Label: "security_opt", Detail: types.CreateStringPointer("array")},
	{Label: "shm_size", Detail: types.CreateStringPointer("number or string")},
	{Label: "stdin_open", Detail: types.CreateStringPointer("boolean or string")},
	{Label: "stop_grace_period", Detail: types.CreateStringPointer("string")},
	{Label: "stop_signal", Detail: types.CreateStringPointer("string")},
	{Label: "storage_opt", Detail: types.CreateStringPointer("object")},
	{Label: "sysctls", Detail: types.CreateStringPointer("array or object")},
	{Label: "tmpfs", Detail: types.CreateStringPointer("array or string")},
	{Label: "tty", Detail: types.CreateStringPointer("boolean or string")},
	{Label: "ulimits", Detail: types.CreateStringPointer("object")},
	{Label: "user", Detail: types.CreateStringPointer("string")},
	{Label: "userns_mode", Detail: types.CreateStringPointer("string")},
	{Label: "uts", Detail: types.CreateStringPointer("string")},
	{Label: "volumes", Detail: types.CreateStringPointer("array")},
	{Label: "volumes_from", Detail: types.CreateStringPointer("array")},
	{Label: "working_dir", Detail: types.CreateStringPointer("string")},
}

func TestCompletion(t *testing.T) {
	testCases := []struct {
		name      string
		content   string
		line      uint32
		character uint32
		list      *protocol.CompletionList
	}{
		{
			name: "top level node suggestions",
			content: `
configs:
  test:
`,
			line:      3,
			character: 0,
			list: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{Label: "configs"},
					{
						Documentation: "compose sub-projects to be included.",
						Label:         "include",
					},
					{
						Documentation: "define the Compose project name, until user defines one explicitly.",
						Label:         "name",
					},
					{Label: "networks"},
					{Label: "secrets"},
					{Label: "services"},
					{
						Documentation: "declared for backward compatibility, ignored.",
						Label:         "version",
					},
					{Label: "volumes"},
				},
			},
		},
		{
			name: "comment prevents suggestions",
			content: `
configs:
  test:
#`,
			line:      3,
			character: 1,
			list:      nil,
		},
		{
			name: "config attributes",
			content: `
configs:
  test:
    `,
			line:      3,
			character: 4,
			list: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{Label: "content", Detail: types.CreateStringPointer("string")},
					{Label: "environment", Detail: types.CreateStringPointer("string")},
					{Label: "external", Detail: types.CreateStringPointer("boolean or string or object")},
					{Label: "file", Detail: types.CreateStringPointer("string")},
					{Label: "labels", Detail: types.CreateStringPointer("array or object")},
					{Label: "name", Detail: types.CreateStringPointer("string")},
					{Label: "template_driver", Detail: types.CreateStringPointer("string")},
				},
			},
		},
		{
			name: "network attributes",
			content: `
networks:
  test:
    `,
			line:      3,
			character: 4,
			list: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{Label: "attachable", Detail: types.CreateStringPointer("boolean or string")},
					{Label: "driver", Detail: types.CreateStringPointer("string")},
					{Label: "driver_opts", Detail: types.CreateStringPointer("object")},
					{Label: "enable_ipv4", Detail: types.CreateStringPointer("boolean or string")},
					{Label: "enable_ipv6", Detail: types.CreateStringPointer("boolean or string")},
					{Label: "external", Detail: types.CreateStringPointer("boolean or string or object")},
					{Label: "internal", Detail: types.CreateStringPointer("boolean or string")},
					{Label: "ipam", Detail: types.CreateStringPointer("object")},
					{Label: "labels", Detail: types.CreateStringPointer("array or object")},
					{Label: "name", Detail: types.CreateStringPointer("string")},
				},
			},
		},
		{
			name: "secret attributes",
			content: `
secrets:
  test:
    `,
			line:      3,
			character: 4,
			list: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{Label: "driver", Detail: types.CreateStringPointer("string")},
					{Label: "driver_opts", Detail: types.CreateStringPointer("object")},
					{Label: "environment", Detail: types.CreateStringPointer("string")},
					{Label: "external", Detail: types.CreateStringPointer("boolean or string or object")},
					{Label: "file", Detail: types.CreateStringPointer("string")},
					{Label: "labels", Detail: types.CreateStringPointer("array or object")},
					{Label: "name", Detail: types.CreateStringPointer("string")},
					{Label: "template_driver", Detail: types.CreateStringPointer("string")},
				},
			},
		},
		{
			name: "service attributes",
			content: `
services:
  test:
    `,
			line:      3,
			character: 4,
			list: &protocol.CompletionList{
				Items: serviceProperties,
			},
		},
		{
			name: "volume attributes",
			content: `
volumes:
  vol:
    `,
			line:      3,
			character: 4,
			list: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{Label: "driver", Detail: types.CreateStringPointer("string")},
					{Label: "driver_opts", Detail: types.CreateStringPointer("object")},
					{Label: "external", Detail: types.CreateStringPointer("boolean or string or object")},
					{Label: "labels", Detail: types.CreateStringPointer("array or object")},
					{Label: "name", Detail: types.CreateStringPointer("string")},
				},
			},
		},
		{
			name: "top level attributes show nothing",
			content: `
configs:
volumes: `,
			line:      2,
			character: 9,
			list:      nil,
		},
		{
			name: "node suggestions do not bleed over to the next top level node",
			content: `
configs:
  configA:
volumes: `,
			line:      3,
			character: 9,
			list:      nil,
		},
		{
			name: "inner attributes of the blkio_config object under service",
			content: `
services:
  postgres:
    blkio_config:
      `,
			line:      4,
			character: 6,
			list: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{Label: "device_read_bps", Detail: types.CreateStringPointer("array")},
					{Label: "device_read_iops", Detail: types.CreateStringPointer("array")},
					{Label: "device_write_bps", Detail: types.CreateStringPointer("array")},
					{Label: "device_write_iops", Detail: types.CreateStringPointer("array")},
					{Label: "weight", Detail: types.CreateStringPointer("integer or string")},
					{Label: "weight_device", Detail: types.CreateStringPointer("array")},
				},
			},
		},
		{
			name: "inner attributes of the deploy object under service",
			content: `
services:
  postgres:
    deploy:
      `,
			line:      4,
			character: 6,
			list: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{Label: "endpoint_mode", Detail: types.CreateStringPointer("string")},
					{Label: "labels", Detail: types.CreateStringPointer("array or object")},
					{Label: "mode", Detail: types.CreateStringPointer("string")},
					{Label: "placement", Detail: types.CreateStringPointer("object")},
					{Label: "replicas", Detail: types.CreateStringPointer("integer or string")},
					{Label: "resources", Detail: types.CreateStringPointer("object")},
					{Label: "restart_policy", Detail: types.CreateStringPointer("object")},
					{Label: "rollback_config", Detail: types.CreateStringPointer("object")},
					{Label: "update_config", Detail: types.CreateStringPointer("object")},
				},
			},
		},
		{
			name: "inner attributes of the deploy/resources object under service",
			content: `
services:
  postgres:
    deploy:
      resources:
        `,
			line:      5,
			character: 8,
			list: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{Label: "limits", Detail: types.CreateStringPointer("object")},
					{Label: "reservations", Detail: types.CreateStringPointer("object")},
				},
			},
		},
		{
			name: "inner attributes of the develop object under service",
			content: `
services:
  postgres:
    develop:
      `,
			line:      4,
			character: 6,
			list: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{Label: "watch", Detail: types.CreateStringPointer("array")},
				},
			},
		},
		{
			name: "inner attributes of the build object under service",
			content: `
services:
  postgres:
    build:
      `,
			line:      4,
			character: 6,
			list: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{Label: "additional_contexts", Detail: types.CreateStringPointer("array or object")},
					{Label: "args", Detail: types.CreateStringPointer("array or object")},
					{Label: "cache_from", Detail: types.CreateStringPointer("array")},
					{Label: "cache_to", Detail: types.CreateStringPointer("array")},
					{Label: "context", Detail: types.CreateStringPointer("string")},
					{Label: "dockerfile", Detail: types.CreateStringPointer("string")},
					{Label: "dockerfile_inline", Detail: types.CreateStringPointer("string")},
					{Label: "entitlements", Detail: types.CreateStringPointer("array")},
					{Label: "extra_hosts", Detail: types.CreateStringPointer("array or object")},
					{Label: "isolation", Detail: types.CreateStringPointer("string")},
					{Label: "labels", Detail: types.CreateStringPointer("array or object")},
					{Label: "network", Detail: types.CreateStringPointer("string")},
					{Label: "no_cache", Detail: types.CreateStringPointer("boolean or string")},
					{Label: "platforms", Detail: types.CreateStringPointer("array")},
					{Label: "privileged", Detail: types.CreateStringPointer("boolean or string")},
					{Label: "pull", Detail: types.CreateStringPointer("boolean or string")},
					{Label: "secrets", Detail: types.CreateStringPointer("array")},
					{Label: "shm_size", Detail: types.CreateStringPointer("integer or string")},
					{Label: "ssh", Detail: types.CreateStringPointer("array or object")},
					{Label: "tags", Detail: types.CreateStringPointer("array")},
					{Label: "target", Detail: types.CreateStringPointer("string")},
					{Label: "ulimits", Detail: types.CreateStringPointer("object")},
				},
			},
		},
		{
			name: "inner attributes of the blkio_config object under service with weight already present",
			content: `
services:
  postgres:
    blkio_config:
      weight: 0
      `,
			line:      5,
			character: 6,
			list: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{Label: "device_read_bps", Detail: types.CreateStringPointer("array")},
					{Label: "device_read_iops", Detail: types.CreateStringPointer("array")},
					{Label: "device_write_bps", Detail: types.CreateStringPointer("array")},
					{Label: "device_write_iops", Detail: types.CreateStringPointer("array")},
					{Label: "weight", Detail: types.CreateStringPointer("integer or string")},
					{Label: "weight_device", Detail: types.CreateStringPointer("array")},
				},
			},
		},
		{
			name: "inner attributes of the ulimits object under service",
			content: `
services:
  test:
    build:
      ulimits:
        abc:
          `,
			line:      6,
			character: 10,
			list: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{Label: "hard", Detail: types.CreateStringPointer("integer or string")},
					{Label: "soft", Detail: types.CreateStringPointer("integer or string")},
				},
			},
		},
		{
			name: "indentation considered for a sibling of a service's attribute",
			content: `
services:
  postgres:
    image: alpine
    `,
			line:      4,
			character: 4,
			list: &protocol.CompletionList{
				Items: serviceProperties,
			},
		},
		{
			name: "indentation considered and does not suggest any items needing a name",
			content: `
services:
  postgres:
    blkio_config:
      weight: 0
  `,
			line:      5,
			character: 2,
			list:      nil,
		},
		{
			name: "indentation considered when suggesting child items",
			content: `
services:
  postgres:
    blkio_config:
      weight: 0
    `,
			line:      5,
			character: 4,
			list: &protocol.CompletionList{
				Items: serviceProperties,
			},
		},
		{
			name: "invalid services as an array",
			content: `
services:
  - `,
			line:      2,
			character: 4,
			list:      nil,
		},
		{
			name: "properties of an embedded object with a custom name",
			content: `
services:
  test:
    networks:
      abc:
        `,
			line:      5,
			character: 8,
			list: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{Label: "aliases", Detail: types.CreateStringPointer("array")},
					{Label: "driver_opts", Detail: types.CreateStringPointer("object")},
					{Label: "gw_priority", Detail: types.CreateStringPointer("number")},
					{Label: "interface_name", Detail: types.CreateStringPointer("string")},
					{Label: "ipv4_address", Detail: types.CreateStringPointer("string")},
					{Label: "ipv6_address", Detail: types.CreateStringPointer("string")},
					{Label: "link_local_ips", Detail: types.CreateStringPointer("array")},
					{Label: "mac_address", Detail: types.CreateStringPointer("string")},
					{Label: "priority", Detail: types.CreateStringPointer("number")},
				},
			},
		},
		{
			name: "oneOf results of a service object's networks attribute",
			content: `
services:
  test:
    networks:
      `,
			line:      4,
			character: 6,
			list:      nil,
		},
		{
			name: "sibling attributes shown after an array of items",
			content: `
services:
  test:
    networks:
    - testNetwork
    `,
			line:      5,
			character: 4,
			list: &protocol.CompletionList{
				Items: serviceProperties,
			},
		},
		{
			name: "properties of a volume array item's sibling attributes under a service object",
			content: `
services:
  test:
    image: alpine
    volumes:
      - type: bind
        `,
			line:      6,
			character: 8,
			list: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{Label: "bind", Detail: types.CreateStringPointer("object")},
					{Label: "consistency", Detail: types.CreateStringPointer("string")},
					{Label: "image", Detail: types.CreateStringPointer("object")},
					{Label: "read_only", Detail: types.CreateStringPointer("boolean or string")},
					{Label: "source", Detail: types.CreateStringPointer("string")},
					{Label: "target", Detail: types.CreateStringPointer("string")},
					{Label: "tmpfs", Detail: types.CreateStringPointer("object")},
					{Label: "type", Detail: types.CreateStringPointer("string")},
					{Label: "volume", Detail: types.CreateStringPointer("object")},
				},
			},
		},
		{
			name: "properties of a volume array item's bind attributes under a service object",
			content: `
services:
  test:
    image: alpine
    volumes:
      - type: bind
        bind:
          `,
			line:      7,
			character: 10,
			list: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{Label: "create_host_path", Detail: types.CreateStringPointer("boolean or string")},
					{Label: "propagation", Detail: types.CreateStringPointer("string")},
					{Label: "recursive", Detail: types.CreateStringPointer("string")},
					{Label: "selinux", Detail: types.CreateStringPointer("string")},
				},
			},
		},
		{
			name: "properties of a volume array item's volume attributes under a service object",
			content: `
services:
  test:
    image: alpine
    volumes:
      - type: bind
        volume:
          `,
			line:      7,
			character: 10,
			list: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{Label: "labels", Detail: types.CreateStringPointer("array or object")},
					{Label: "nocopy", Detail: types.CreateStringPointer("boolean or string")},
					{Label: "subpath", Detail: types.CreateStringPointer("string")},
				},
			},
		},
		{
			name: "enum properties for a string attribute is suggested",
			content: `
services:
  test:
    volumes:
      - type: bind
        bind:
          selinux: `,
			line:      6,
			character: 19,
			list: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{Label: "Z", Detail: types.CreateStringPointer("string")},
					{Label: "z", Detail: types.CreateStringPointer("string")},
				},
			},
		},
		{
			name: "nothing suggested for an arbitrary string attribute's value",
			content: `
services:
  test:
    volumes:
      - type: bind
        bind:
          propagation:`,
			line:      6,
			character: 22,
			list:      nil,
		},
	}

	composeFileURI := fmt.Sprintf("file:///%v", strings.TrimPrefix(filepath.ToSlash(filepath.Join(os.TempDir(), "compose.yaml")), "/"))

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc := document.NewComposeDocument(uri.URI(composeFileURI), 1, []byte(tc.content))
			list, err := Completion(context.Background(), &protocol.CompletionParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{URI: composeFileURI},
					Position:     protocol.Position{Line: tc.line, Character: tc.character},
				},
			}, doc)
			require.NoError(t, err)
			if tc.list == nil {
				require.Nil(t, list)
			} else {
				require.Equal(t, tc.list, list)
			}
		})
	}
}
