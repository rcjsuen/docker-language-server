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
	"github.com/stretchr/testify/require"
	"go.lsp.dev/uri"
)

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
					{Label: "content"},
					{Label: "environment"},
					{Label: "external"},
					{Label: "file"},
					{Label: "labels"},
					{Label: "name"},
					{Label: "template_driver"},
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
					{Label: "attachable"},
					{Label: "driver"},
					{Label: "driver_opts"},
					{Label: "enable_ipv4"},
					{Label: "enable_ipv6"},
					{Label: "external"},
					{Label: "internal"},
					{Label: "ipam"},
					{Label: "labels"},
					{Label: "name"},
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
					{Label: "driver"},
					{Label: "driver_opts"},
					{Label: "environment"},
					{Label: "external"},
					{Label: "file"},
					{Label: "labels"},
					{Label: "name"},
					{Label: "template_driver"},
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
				Items: []protocol.CompletionItem{
					{Label: "annotations"},
					{Label: "attach"},
					{Label: "blkio_config"},
					{Label: "build"},
					{Label: "cap_add"},
					{Label: "cap_drop"},
					{Label: "cgroup"},
					{Label: "cgroup_parent"},
					{Label: "command"},
					{Label: "configs"},
					{Label: "container_name"},
					{Label: "cpu_count"},
					{Label: "cpu_percent"},
					{Label: "cpu_period"},
					{Label: "cpu_quota"},
					{Label: "cpu_rt_period"},
					{Label: "cpu_rt_runtime"},
					{Label: "cpu_shares"},
					{Label: "cpus"},
					{Label: "cpuset"},
					{Label: "credential_spec"},
					{Label: "depends_on"},
					{Label: "deploy"},
					{Label: "develop"},
					{Label: "device_cgroup_rules"},
					{Label: "devices"},
					{Label: "dns"},
					{Label: "dns_opt"},
					{Label: "dns_search"},
					{Label: "domainname"},
					{Label: "entrypoint"},
					{Label: "env_file"},
					{Label: "environment"},
					{Label: "expose"},
					{Label: "extends"},
					{Label: "external_links"},
					{Label: "extra_hosts"},
					{Label: "gpus"},
					{Label: "group_add"},
					{Label: "healthcheck"},
					{Label: "hostname"},
					{Label: "image"},
					{Label: "init"},
					{Label: "ipc"},
					{Label: "isolation"},
					{Label: "label_file"},
					{Label: "labels"},
					{Label: "links"},
					{Label: "logging"},
					{Label: "mac_address"},
					{Label: "mem_limit"},
					{Label: "mem_reservation"},
					{Label: "mem_swappiness"},
					{Label: "memswap_limit"},
					{Label: "network_mode"},
					{Label: "networks"},
					{Label: "oom_kill_disable"},
					{Label: "oom_score_adj"},
					{Label: "pid"},
					{Label: "pids_limit"},
					{Label: "platform"},
					{Label: "ports"},
					{Label: "post_start"},
					{Label: "pre_stop"},
					{Label: "privileged"},
					{Label: "profiles"},
					{Label: "provider"},
					{Label: "pull_policy"},
					{Label: "pull_refresh_after"},
					{Label: "read_only"},
					{Label: "restart"},
					{Label: "runtime"},
					{Label: "scale"},
					{Label: "secrets"},
					{Label: "security_opt"},
					{Label: "shm_size"},
					{Label: "stdin_open"},
					{Label: "stop_grace_period"},
					{Label: "stop_signal"},
					{Label: "storage_opt"},
					{Label: "sysctls"},
					{Label: "tmpfs"},
					{Label: "tty"},
					{Label: "ulimits"},
					{Label: "user"},
					{Label: "userns_mode"},
					{Label: "uts"},
					{Label: "volumes"},
					{Label: "volumes_from"},
					{Label: "working_dir"},
				},
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
					{Label: "driver"},
					{Label: "driver_opts"},
					{Label: "external"},
					{Label: "labels"},
					{Label: "name"},
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
					{Label: "device_read_bps"},
					{Label: "device_read_iops"},
					{Label: "device_write_bps"},
					{Label: "device_write_iops"},
					{Label: "weight"},
					{Label: "weight_device"},
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
					{Label: "endpoint_mode"},
					{Label: "labels"},
					{Label: "mode"},
					{Label: "placement"},
					{Label: "replicas"},
					{Label: "resources"},
					{Label: "restart_policy"},
					{Label: "rollback_config"},
					{Label: "update_config"},
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
					{Label: "limits"},
					{Label: "reservations"},
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
					{Label: "watch"},
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
					{Label: "additional_contexts"},
					{Label: "args"},
					{Label: "cache_from"},
					{Label: "cache_to"},
					{Label: "context"},
					{Label: "dockerfile"},
					{Label: "dockerfile_inline"},
					{Label: "entitlements"},
					{Label: "extra_hosts"},
					{Label: "isolation"},
					{Label: "labels"},
					{Label: "network"},
					{Label: "no_cache"},
					{Label: "platforms"},
					{Label: "privileged"},
					{Label: "pull"},
					{Label: "secrets"},
					{Label: "shm_size"},
					{Label: "ssh"},
					{Label: "tags"},
					{Label: "target"},
					{Label: "ulimits"},
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
					{Label: "device_read_bps"},
					{Label: "device_read_iops"},
					{Label: "device_write_bps"},
					{Label: "device_write_iops"},
					{Label: "weight"},
					{Label: "weight_device"},
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
					{Label: "hard"},
					{Label: "soft"},
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
				Items: []protocol.CompletionItem{
					{Label: "annotations"},
					{Label: "attach"},
					{Label: "blkio_config"},
					{Label: "build"},
					{Label: "cap_add"},
					{Label: "cap_drop"},
					{Label: "cgroup"},
					{Label: "cgroup_parent"},
					{Label: "command"},
					{Label: "configs"},
					{Label: "container_name"},
					{Label: "cpu_count"},
					{Label: "cpu_percent"},
					{Label: "cpu_period"},
					{Label: "cpu_quota"},
					{Label: "cpu_rt_period"},
					{Label: "cpu_rt_runtime"},
					{Label: "cpu_shares"},
					{Label: "cpus"},
					{Label: "cpuset"},
					{Label: "credential_spec"},
					{Label: "depends_on"},
					{Label: "deploy"},
					{Label: "develop"},
					{Label: "device_cgroup_rules"},
					{Label: "devices"},
					{Label: "dns"},
					{Label: "dns_opt"},
					{Label: "dns_search"},
					{Label: "domainname"},
					{Label: "entrypoint"},
					{Label: "env_file"},
					{Label: "environment"},
					{Label: "expose"},
					{Label: "extends"},
					{Label: "external_links"},
					{Label: "extra_hosts"},
					{Label: "gpus"},
					{Label: "group_add"},
					{Label: "healthcheck"},
					{Label: "hostname"},
					{Label: "image"},
					{Label: "init"},
					{Label: "ipc"},
					{Label: "isolation"},
					{Label: "label_file"},
					{Label: "labels"},
					{Label: "links"},
					{Label: "logging"},
					{Label: "mac_address"},
					{Label: "mem_limit"},
					{Label: "mem_reservation"},
					{Label: "mem_swappiness"},
					{Label: "memswap_limit"},
					{Label: "network_mode"},
					{Label: "networks"},
					{Label: "oom_kill_disable"},
					{Label: "oom_score_adj"},
					{Label: "pid"},
					{Label: "pids_limit"},
					{Label: "platform"},
					{Label: "ports"},
					{Label: "post_start"},
					{Label: "pre_stop"},
					{Label: "privileged"},
					{Label: "profiles"},
					{Label: "provider"},
					{Label: "pull_policy"},
					{Label: "pull_refresh_after"},
					{Label: "read_only"},
					{Label: "restart"},
					{Label: "runtime"},
					{Label: "scale"},
					{Label: "secrets"},
					{Label: "security_opt"},
					{Label: "shm_size"},
					{Label: "stdin_open"},
					{Label: "stop_grace_period"},
					{Label: "stop_signal"},
					{Label: "storage_opt"},
					{Label: "sysctls"},
					{Label: "tmpfs"},
					{Label: "tty"},
					{Label: "ulimits"},
					{Label: "user"},
					{Label: "userns_mode"},
					{Label: "uts"},
					{Label: "volumes"},
					{Label: "volumes_from"},
					{Label: "working_dir"},
				},
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
				Items: []protocol.CompletionItem{
					{Label: "annotations"},
					{Label: "attach"},
					{Label: "blkio_config"},
					{Label: "build"},
					{Label: "cap_add"},
					{Label: "cap_drop"},
					{Label: "cgroup"},
					{Label: "cgroup_parent"},
					{Label: "command"},
					{Label: "configs"},
					{Label: "container_name"},
					{Label: "cpu_count"},
					{Label: "cpu_percent"},
					{Label: "cpu_period"},
					{Label: "cpu_quota"},
					{Label: "cpu_rt_period"},
					{Label: "cpu_rt_runtime"},
					{Label: "cpu_shares"},
					{Label: "cpus"},
					{Label: "cpuset"},
					{Label: "credential_spec"},
					{Label: "depends_on"},
					{Label: "deploy"},
					{Label: "develop"},
					{Label: "device_cgroup_rules"},
					{Label: "devices"},
					{Label: "dns"},
					{Label: "dns_opt"},
					{Label: "dns_search"},
					{Label: "domainname"},
					{Label: "entrypoint"},
					{Label: "env_file"},
					{Label: "environment"},
					{Label: "expose"},
					{Label: "extends"},
					{Label: "external_links"},
					{Label: "extra_hosts"},
					{Label: "gpus"},
					{Label: "group_add"},
					{Label: "healthcheck"},
					{Label: "hostname"},
					{Label: "image"},
					{Label: "init"},
					{Label: "ipc"},
					{Label: "isolation"},
					{Label: "label_file"},
					{Label: "labels"},
					{Label: "links"},
					{Label: "logging"},
					{Label: "mac_address"},
					{Label: "mem_limit"},
					{Label: "mem_reservation"},
					{Label: "mem_swappiness"},
					{Label: "memswap_limit"},
					{Label: "network_mode"},
					{Label: "networks"},
					{Label: "oom_kill_disable"},
					{Label: "oom_score_adj"},
					{Label: "pid"},
					{Label: "pids_limit"},
					{Label: "platform"},
					{Label: "ports"},
					{Label: "post_start"},
					{Label: "pre_stop"},
					{Label: "privileged"},
					{Label: "profiles"},
					{Label: "provider"},
					{Label: "pull_policy"},
					{Label: "pull_refresh_after"},
					{Label: "read_only"},
					{Label: "restart"},
					{Label: "runtime"},
					{Label: "scale"},
					{Label: "secrets"},
					{Label: "security_opt"},
					{Label: "shm_size"},
					{Label: "stdin_open"},
					{Label: "stop_grace_period"},
					{Label: "stop_signal"},
					{Label: "storage_opt"},
					{Label: "sysctls"},
					{Label: "tmpfs"},
					{Label: "tty"},
					{Label: "ulimits"},
					{Label: "user"},
					{Label: "userns_mode"},
					{Label: "uts"},
					{Label: "volumes"},
					{Label: "volumes_from"},
					{Label: "working_dir"},
				},
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
					{Label: "aliases"},
					{Label: "driver_opts"},
					{Label: "gw_priority"},
					{Label: "interface_name"},
					{Label: "ipv4_address"},
					{Label: "ipv6_address"},
					{Label: "link_local_ips"},
					{Label: "mac_address"},
					{Label: "priority"},
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
				Items: []protocol.CompletionItem{
					{Label: "annotations"},
					{Label: "attach"},
					{Label: "blkio_config"},
					{Label: "build"},
					{Label: "cap_add"},
					{Label: "cap_drop"},
					{Label: "cgroup"},
					{Label: "cgroup_parent"},
					{Label: "command"},
					{Label: "configs"},
					{Label: "container_name"},
					{Label: "cpu_count"},
					{Label: "cpu_percent"},
					{Label: "cpu_period"},
					{Label: "cpu_quota"},
					{Label: "cpu_rt_period"},
					{Label: "cpu_rt_runtime"},
					{Label: "cpu_shares"},
					{Label: "cpus"},
					{Label: "cpuset"},
					{Label: "credential_spec"},
					{Label: "depends_on"},
					{Label: "deploy"},
					{Label: "develop"},
					{Label: "device_cgroup_rules"},
					{Label: "devices"},
					{Label: "dns"},
					{Label: "dns_opt"},
					{Label: "dns_search"},
					{Label: "domainname"},
					{Label: "entrypoint"},
					{Label: "env_file"},
					{Label: "environment"},
					{Label: "expose"},
					{Label: "extends"},
					{Label: "external_links"},
					{Label: "extra_hosts"},
					{Label: "gpus"},
					{Label: "group_add"},
					{Label: "healthcheck"},
					{Label: "hostname"},
					{Label: "image"},
					{Label: "init"},
					{Label: "ipc"},
					{Label: "isolation"},
					{Label: "label_file"},
					{Label: "labels"},
					{Label: "links"},
					{Label: "logging"},
					{Label: "mac_address"},
					{Label: "mem_limit"},
					{Label: "mem_reservation"},
					{Label: "mem_swappiness"},
					{Label: "memswap_limit"},
					{Label: "network_mode"},
					{Label: "networks"},
					{Label: "oom_kill_disable"},
					{Label: "oom_score_adj"},
					{Label: "pid"},
					{Label: "pids_limit"},
					{Label: "platform"},
					{Label: "ports"},
					{Label: "post_start"},
					{Label: "pre_stop"},
					{Label: "privileged"},
					{Label: "profiles"},
					{Label: "provider"},
					{Label: "pull_policy"},
					{Label: "pull_refresh_after"},
					{Label: "read_only"},
					{Label: "restart"},
					{Label: "runtime"},
					{Label: "scale"},
					{Label: "secrets"},
					{Label: "security_opt"},
					{Label: "shm_size"},
					{Label: "stdin_open"},
					{Label: "stop_grace_period"},
					{Label: "stop_signal"},
					{Label: "storage_opt"},
					{Label: "sysctls"},
					{Label: "tmpfs"},
					{Label: "tty"},
					{Label: "ulimits"},
					{Label: "user"},
					{Label: "userns_mode"},
					{Label: "uts"},
					{Label: "volumes"},
					{Label: "volumes_from"},
					{Label: "working_dir"},
				},
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
					{Label: "bind"},
					{Label: "consistency"},
					{Label: "image"},
					{Label: "read_only"},
					{Label: "source"},
					{Label: "target"},
					{Label: "tmpfs"},
					{Label: "type"},
					{Label: "volume"},
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
					{Label: "create_host_path"},
					{Label: "propagation"},
					{Label: "recursive"},
					{Label: "selinux"},
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
					{Label: "labels"},
					{Label: "nocopy"},
					{Label: "subpath"},
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
					{Label: "Z"},
					{Label: "z"},
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
