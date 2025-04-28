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
	{
		Label:          "annotations",
		Detail:         types.CreateStringPointer("array or object"),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "attach",
		Detail:         types.CreateStringPointer("boolean or string"),
		InsertText:     types.CreateStringPointer("attach: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "blkio_config",
		Detail:         types.CreateStringPointer("object"),
		InsertText:     types.CreateStringPointer("blkio_config:\n      "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "build",
		Detail:         types.CreateStringPointer("object or string"),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "cap_add",
		Detail:         types.CreateStringPointer("array"),
		InsertText:     types.CreateStringPointer("cap_add:\n      - "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "cap_drop",
		Detail:         types.CreateStringPointer("array"),
		InsertText:     types.CreateStringPointer("cap_drop:\n      - "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:            "cgroup",
		Detail:           types.CreateStringPointer("string"),
		InsertText:       types.CreateStringPointer("cgroup: ${1|host,private|}"),
		InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "cgroup_parent",
		Detail:         types.CreateStringPointer("string"),
		InsertText:     types.CreateStringPointer("cgroup_parent: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "command",
		Detail:         types.CreateStringPointer("array or null or string"),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "configs",
		Detail:         types.CreateStringPointer("array"),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "container_name",
		Detail:         types.CreateStringPointer("string"),
		InsertText:     types.CreateStringPointer("container_name: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "cpu_count",
		Detail:         types.CreateStringPointer("integer or string"),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "cpu_percent",
		Detail:         types.CreateStringPointer("integer or string"),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "cpu_period",
		Detail:         types.CreateStringPointer("number or string"),
		InsertText:     types.CreateStringPointer("cpu_period: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "cpu_quota",
		Detail:         types.CreateStringPointer("number or string"),
		InsertText:     types.CreateStringPointer("cpu_quota: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "cpu_rt_period",
		Detail:         types.CreateStringPointer("number or string"),
		InsertText:     types.CreateStringPointer("cpu_rt_period: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "cpu_rt_runtime",
		Detail:         types.CreateStringPointer("number or string"),
		InsertText:     types.CreateStringPointer("cpu_rt_runtime: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "cpu_shares",
		Detail:         types.CreateStringPointer("number or string"),
		InsertText:     types.CreateStringPointer("cpu_shares: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "cpus",
		Detail:         types.CreateStringPointer("number or string"),
		InsertText:     types.CreateStringPointer("cpus: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "cpuset",
		Detail:         types.CreateStringPointer("string"),
		InsertText:     types.CreateStringPointer("cpuset: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "credential_spec",
		Detail:         types.CreateStringPointer("object"),
		InsertText:     types.CreateStringPointer("credential_spec:\n      "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "depends_on",
		Detail:         types.CreateStringPointer("array or object"),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "deploy",
		Detail:         types.CreateStringPointer("null or object"),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "develop",
		Detail:         types.CreateStringPointer("null or object"),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "device_cgroup_rules",
		Detail:         types.CreateStringPointer("array"),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "devices",
		Detail:         types.CreateStringPointer("array"),
		InsertText:     types.CreateStringPointer("devices:\n      - "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "dns",
		Detail:         types.CreateStringPointer("array or string"),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "dns_opt",
		Detail:         types.CreateStringPointer("array"),
		InsertText:     types.CreateStringPointer("dns_opt:\n      - "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "dns_search",
		Detail:         types.CreateStringPointer("array or string"),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "domainname",
		Detail:         types.CreateStringPointer("string"),
		InsertText:     types.CreateStringPointer("domainname: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "entrypoint",
		Detail:         types.CreateStringPointer("array or null or string"),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "env_file",
		Detail:         types.CreateStringPointer("array or string"),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "environment",
		Detail:         types.CreateStringPointer("array or object"),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "expose",
		Detail:         types.CreateStringPointer("array"),
		InsertText:     types.CreateStringPointer("expose:\n      - "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "extends",
		Detail:         types.CreateStringPointer("object or string"),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "external_links",
		Detail:         types.CreateStringPointer("array"),
		InsertText:     types.CreateStringPointer("external_links:\n      - "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "extra_hosts",
		Detail:         types.CreateStringPointer("array or object"),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "gpus",
		Detail:         types.CreateStringPointer("array or string"),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "group_add",
		Detail:         types.CreateStringPointer("array"),
		InsertText:     types.CreateStringPointer("group_add:\n      - "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "healthcheck",
		Detail:         types.CreateStringPointer("object"),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "hostname",
		Detail:         types.CreateStringPointer("string"),
		InsertText:     types.CreateStringPointer("hostname: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "image",
		Detail:         types.CreateStringPointer("string"),
		InsertText:     types.CreateStringPointer("image: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "init",
		Detail:         types.CreateStringPointer("boolean or string"),
		InsertText:     types.CreateStringPointer("init: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "ipc",
		Detail:         types.CreateStringPointer("string"),
		InsertText:     types.CreateStringPointer("ipc: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "isolation",
		Detail:         types.CreateStringPointer("string"),
		InsertText:     types.CreateStringPointer("isolation: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "label_file",
		Detail:         types.CreateStringPointer("array or string"),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "labels",
		Detail:         types.CreateStringPointer("array or object"),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "links",
		Detail:         types.CreateStringPointer("array"),
		InsertText:     types.CreateStringPointer("links:\n      - "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "logging",
		Detail:         types.CreateStringPointer("object"),
		InsertText:     types.CreateStringPointer("logging:\n      "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "mac_address",
		Detail:         types.CreateStringPointer("string"),
		InsertText:     types.CreateStringPointer("mac_address: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "mem_limit",
		Detail:         types.CreateStringPointer("number or string"),
		InsertText:     types.CreateStringPointer("mem_limit: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "mem_reservation",
		Detail:         types.CreateStringPointer("integer or string"),
		InsertText:     types.CreateStringPointer("mem_reservation: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "mem_swappiness",
		Detail:         types.CreateStringPointer("integer or string"),
		InsertText:     types.CreateStringPointer("mem_swappiness: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "memswap_limit",
		Detail:         types.CreateStringPointer("number or string"),
		InsertText:     types.CreateStringPointer("memswap_limit: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "network_mode",
		Detail:         types.CreateStringPointer("string"),
		InsertText:     types.CreateStringPointer("network_mode: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "networks",
		Detail:         types.CreateStringPointer("array or object"),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "oom_kill_disable",
		Detail:         types.CreateStringPointer("boolean or string"),
		InsertText:     types.CreateStringPointer("oom_kill_disable: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "oom_score_adj",
		Detail:         types.CreateStringPointer("integer or string"),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "pid",
		Detail:         types.CreateStringPointer("null or string"),
		InsertText:     types.CreateStringPointer("pid: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "pids_limit",
		Detail:         types.CreateStringPointer("number or string"),
		InsertText:     types.CreateStringPointer("pids_limit: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "platform",
		Detail:         types.CreateStringPointer("string"),
		InsertText:     types.CreateStringPointer("platform: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "ports",
		Detail:         types.CreateStringPointer("array"),
		InsertText:     types.CreateStringPointer("ports:\n      - "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "post_start",
		Detail:         types.CreateStringPointer("array"),
		InsertText:     types.CreateStringPointer("post_start:\n      - "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "pre_stop",
		Detail:         types.CreateStringPointer("array"),
		InsertText:     types.CreateStringPointer("pre_stop:\n      - "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "privileged",
		Detail:         types.CreateStringPointer("boolean or string"),
		InsertText:     types.CreateStringPointer("privileged: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "profiles",
		Detail:         types.CreateStringPointer("array"),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "provider",
		Detail:         types.CreateStringPointer("object"),
		InsertText:     types.CreateStringPointer("provider:\n      "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "pull_policy",
		Detail:         types.CreateStringPointer("string"),
		InsertText:     types.CreateStringPointer("pull_policy: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "pull_refresh_after",
		Detail:         types.CreateStringPointer("string"),
		InsertText:     types.CreateStringPointer("pull_refresh_after: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "read_only",
		Detail:         types.CreateStringPointer("boolean or string"),
		InsertText:     types.CreateStringPointer("read_only: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "restart",
		Detail:         types.CreateStringPointer("string"),
		InsertText:     types.CreateStringPointer("restart: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "runtime",
		Detail:         types.CreateStringPointer("string"),
		InsertText:     types.CreateStringPointer("runtime: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "scale",
		Detail:         types.CreateStringPointer("integer or string"),
		InsertText:     types.CreateStringPointer("scale: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "secrets",
		Detail:         types.CreateStringPointer("array"),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "security_opt",
		Detail:         types.CreateStringPointer("array"),
		InsertText:     types.CreateStringPointer("security_opt:\n      - "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "shm_size",
		Detail:         types.CreateStringPointer("number or string"),
		InsertText:     types.CreateStringPointer("shm_size: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "stdin_open",
		Detail:         types.CreateStringPointer("boolean or string"),
		InsertText:     types.CreateStringPointer("stdin_open: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "stop_grace_period",
		Detail:         types.CreateStringPointer("string"),
		InsertText:     types.CreateStringPointer("stop_grace_period: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "stop_signal",
		Detail:         types.CreateStringPointer("string"),
		InsertText:     types.CreateStringPointer("stop_signal: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "storage_opt",
		Detail:         types.CreateStringPointer("object"),
		InsertText:     types.CreateStringPointer("storage_opt:\n      "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "sysctls",
		Detail:         types.CreateStringPointer("array or object"),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "tmpfs",
		Detail:         types.CreateStringPointer("array or string"),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "tty",
		Detail:         types.CreateStringPointer("boolean or string"),
		InsertText:     types.CreateStringPointer("tty: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "ulimits",
		Detail:         types.CreateStringPointer("object"),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "user",
		Detail:         types.CreateStringPointer("string"),
		InsertText:     types.CreateStringPointer("user: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "userns_mode",
		Detail:         types.CreateStringPointer("string"),
		InsertText:     types.CreateStringPointer("userns_mode: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "uts",
		Detail:         types.CreateStringPointer("string"),
		InsertText:     types.CreateStringPointer("uts: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "volumes",
		Detail:         types.CreateStringPointer("array"),
		InsertText:     types.CreateStringPointer("volumes:\n      - "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "volumes_from",
		Detail:         types.CreateStringPointer("array"),
		InsertText:     types.CreateStringPointer("volumes_from:\n      - "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
	{
		Label:          "working_dir",
		Detail:         types.CreateStringPointer("string"),
		InsertText:     types.CreateStringPointer("working_dir: "),
		InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
	},
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
					{
						Label:          "content",
						Detail:         types.CreateStringPointer("string"),
						InsertText:     types.CreateStringPointer("content: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "environment",
						Detail:         types.CreateStringPointer("string"),
						InsertText:     types.CreateStringPointer("environment: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "external",
						Detail:         types.CreateStringPointer("boolean or string or object"),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "file",
						Detail:         types.CreateStringPointer("string"),
						InsertText:     types.CreateStringPointer("file: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "labels",
						Detail:         types.CreateStringPointer("array or object"),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "name",
						Detail:         types.CreateStringPointer("string"),
						InsertText:     types.CreateStringPointer("name: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "template_driver",
						Detail:         types.CreateStringPointer("string"),
						InsertText:     types.CreateStringPointer("template_driver: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
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
					{
						Label:          "attachable",
						Detail:         types.CreateStringPointer("boolean or string"),
						InsertText:     types.CreateStringPointer("attachable: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "driver",
						Detail:         types.CreateStringPointer("string"),
						InsertText:     types.CreateStringPointer("driver: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "driver_opts",
						Detail:         types.CreateStringPointer("object"),
						InsertText:     types.CreateStringPointer("driver_opts:\n      "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "enable_ipv4",
						Detail:         types.CreateStringPointer("boolean or string"),
						InsertText:     types.CreateStringPointer("enable_ipv4: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "enable_ipv6",
						Detail:         types.CreateStringPointer("boolean or string"),
						InsertText:     types.CreateStringPointer("enable_ipv6: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "external",
						Detail:         types.CreateStringPointer("boolean or string or object"),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "internal",
						Detail:         types.CreateStringPointer("boolean or string"),
						InsertText:     types.CreateStringPointer("internal: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "ipam",
						Detail:         types.CreateStringPointer("object"),
						InsertText:     types.CreateStringPointer("ipam:\n      "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "labels",
						Detail:         types.CreateStringPointer("array or object"),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "name",
						Detail:         types.CreateStringPointer("string"),
						InsertText:     types.CreateStringPointer("name: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
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
					{
						Label:          "driver",
						Detail:         types.CreateStringPointer("string"),
						InsertText:     types.CreateStringPointer("driver: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "driver_opts",
						Detail:         types.CreateStringPointer("object"),
						InsertText:     types.CreateStringPointer("driver_opts:\n      "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "environment",
						Detail:         types.CreateStringPointer("string"),
						InsertText:     types.CreateStringPointer("environment: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "external",
						Detail:         types.CreateStringPointer("boolean or string or object"),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "file",
						Detail:         types.CreateStringPointer("string"),
						InsertText:     types.CreateStringPointer("file: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "labels",
						Detail:         types.CreateStringPointer("array or object"),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "name",
						Detail:         types.CreateStringPointer("string"),
						InsertText:     types.CreateStringPointer("name: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "template_driver",
						Detail:         types.CreateStringPointer("string"),
						InsertText:     types.CreateStringPointer("template_driver: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
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
					{
						Label:          "driver",
						Detail:         types.CreateStringPointer("string"),
						InsertText:     types.CreateStringPointer("driver: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "driver_opts",
						Detail:         types.CreateStringPointer("object"),
						InsertText:     types.CreateStringPointer("driver_opts:\n      "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "external",
						Detail:         types.CreateStringPointer("boolean or string or object"),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "labels",
						Detail:         types.CreateStringPointer("array or object"),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "name",
						Detail:         types.CreateStringPointer("string"),
						InsertText:     types.CreateStringPointer("name: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
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
					{
						Label:          "device_read_bps",
						Detail:         types.CreateStringPointer("array"),
						InsertText:     types.CreateStringPointer("device_read_bps:\n        - "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "device_read_iops",
						Detail:         types.CreateStringPointer("array"),
						InsertText:     types.CreateStringPointer("device_read_iops:\n        - "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "device_write_bps",
						Detail:         types.CreateStringPointer("array"),
						InsertText:     types.CreateStringPointer("device_write_bps:\n        - "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "device_write_iops",
						Detail:         types.CreateStringPointer("array"),
						InsertText:     types.CreateStringPointer("device_write_iops:\n        - "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "weight",
						Detail:         types.CreateStringPointer("integer or string"),
						InsertText:     types.CreateStringPointer("weight: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "weight_device",
						Detail:         types.CreateStringPointer("array"),
						InsertText:     types.CreateStringPointer("weight_device:\n        - "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
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
					{
						Label:          "endpoint_mode",
						Detail:         types.CreateStringPointer("string"),
						InsertText:     types.CreateStringPointer("endpoint_mode: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "labels",
						Detail:         types.CreateStringPointer("array or object"),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "mode",
						Detail:         types.CreateStringPointer("string"),
						InsertText:     types.CreateStringPointer("mode: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "placement",
						Detail:         types.CreateStringPointer("object"),
						InsertText:     types.CreateStringPointer("placement:\n        "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "replicas",
						Detail:         types.CreateStringPointer("integer or string"),
						InsertText:     types.CreateStringPointer("replicas: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "resources",
						Detail:         types.CreateStringPointer("object"),
						InsertText:     types.CreateStringPointer("resources:\n        "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "restart_policy",
						Detail:         types.CreateStringPointer("object"),
						InsertText:     types.CreateStringPointer("restart_policy:\n        "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "rollback_config",
						Detail:         types.CreateStringPointer("object"),
						InsertText:     types.CreateStringPointer("rollback_config:\n        "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "update_config",
						Detail:         types.CreateStringPointer("object"),
						InsertText:     types.CreateStringPointer("update_config:\n        "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
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
					{
						Label:          "limits",
						Detail:         types.CreateStringPointer("object"),
						InsertText:     types.CreateStringPointer("limits:\n          "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "reservations",
						Detail:         types.CreateStringPointer("object"),
						InsertText:     types.CreateStringPointer("reservations:\n          "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
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
					{
						Label:          "watch",
						Detail:         types.CreateStringPointer("array"),
						InsertText:     types.CreateStringPointer("watch:\n        - "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
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
					{
						Label:          "additional_contexts",
						Detail:         types.CreateStringPointer("array or object"),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "args",
						Detail:         types.CreateStringPointer("array or object"),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "cache_from",
						Detail:         types.CreateStringPointer("array"),
						InsertText:     types.CreateStringPointer("cache_from:\n        - "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "cache_to",
						Detail:         types.CreateStringPointer("array"),
						InsertText:     types.CreateStringPointer("cache_to:\n        - "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "context",
						Detail:         types.CreateStringPointer("string"),
						InsertText:     types.CreateStringPointer("context: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "dockerfile",
						Detail:         types.CreateStringPointer("string"),
						InsertText:     types.CreateStringPointer("dockerfile: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "dockerfile_inline",
						Detail:         types.CreateStringPointer("string"),
						InsertText:     types.CreateStringPointer("dockerfile_inline: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "entitlements",
						Detail:         types.CreateStringPointer("array"),
						InsertText:     types.CreateStringPointer("entitlements:\n        - "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "extra_hosts",
						Detail:         types.CreateStringPointer("array or object"),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "isolation",
						Detail:         types.CreateStringPointer("string"),
						InsertText:     types.CreateStringPointer("isolation: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "labels",
						Detail:         types.CreateStringPointer("array or object"),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "network",
						Detail:         types.CreateStringPointer("string"),
						InsertText:     types.CreateStringPointer("network: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "no_cache",
						Detail:         types.CreateStringPointer("boolean or string"),
						InsertText:     types.CreateStringPointer("no_cache: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "platforms",
						Detail:         types.CreateStringPointer("array"),
						InsertText:     types.CreateStringPointer("platforms:\n        - "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "privileged",
						Detail:         types.CreateStringPointer("boolean or string"),
						InsertText:     types.CreateStringPointer("privileged: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "pull",
						Detail:         types.CreateStringPointer("boolean or string"),
						InsertText:     types.CreateStringPointer("pull: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "secrets",
						Detail:         types.CreateStringPointer("array"),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "shm_size",
						Detail:         types.CreateStringPointer("integer or string"),
						InsertText:     types.CreateStringPointer("shm_size: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "ssh",
						Detail:         types.CreateStringPointer("array or object"),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "tags",
						Detail:         types.CreateStringPointer("array"),
						InsertText:     types.CreateStringPointer("tags:\n        - "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "target",
						Detail:         types.CreateStringPointer("string"),
						InsertText:     types.CreateStringPointer("target: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label: "ulimits", Detail: types.CreateStringPointer("object"),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
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
					{
						Label:          "device_read_bps",
						Detail:         types.CreateStringPointer("array"),
						InsertText:     types.CreateStringPointer("device_read_bps:\n        - "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "device_read_iops",
						Detail:         types.CreateStringPointer("array"),
						InsertText:     types.CreateStringPointer("device_read_iops:\n        - "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "device_write_bps",
						Detail:         types.CreateStringPointer("array"),
						InsertText:     types.CreateStringPointer("device_write_bps:\n        - "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "device_write_iops",
						Detail:         types.CreateStringPointer("array"),
						InsertText:     types.CreateStringPointer("device_write_iops:\n        - "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "weight",
						Detail:         types.CreateStringPointer("integer or string"),
						InsertText:     types.CreateStringPointer("weight: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "weight_device",
						Detail:         types.CreateStringPointer("array"),
						InsertText:     types.CreateStringPointer("weight_device:\n        - "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
				},
			},
		},
		{
			name: "inner attributes of the configs array object under service",
			content: `
services:
  test:
    configs:
    - `,
			line:      4,
			character: 6,
			list: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{
						Label:          "gid",
						Detail:         types.CreateStringPointer("string"),
						InsertText:     types.CreateStringPointer("gid: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "mode",
						Detail:         types.CreateStringPointer("number or string"),
						InsertText:     types.CreateStringPointer("mode: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "source",
						Detail:         types.CreateStringPointer("string"),
						InsertText:     types.CreateStringPointer("source: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "target",
						Detail:         types.CreateStringPointer("string"),
						InsertText:     types.CreateStringPointer("target: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "uid",
						Detail:         types.CreateStringPointer("string"),
						InsertText:     types.CreateStringPointer("uid: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
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
					{
						Label:          "hard",
						Detail:         types.CreateStringPointer("integer or string"),
						InsertText:     types.CreateStringPointer("hard: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "soft",
						Detail:         types.CreateStringPointer("integer or string"),
						InsertText:     types.CreateStringPointer("soft: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
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
					{
						Label:          "aliases",
						Detail:         types.CreateStringPointer("array"),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "driver_opts",
						Detail:         types.CreateStringPointer("object"),
						InsertText:     types.CreateStringPointer("driver_opts:\n          "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "gw_priority",
						Detail:         types.CreateStringPointer("number"),
						InsertText:     types.CreateStringPointer("gw_priority: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "interface_name",
						Detail:         types.CreateStringPointer("string"),
						InsertText:     types.CreateStringPointer("interface_name: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "ipv4_address",
						Detail:         types.CreateStringPointer("string"),
						InsertText:     types.CreateStringPointer("ipv4_address: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "ipv6_address",
						Detail:         types.CreateStringPointer("string"),
						InsertText:     types.CreateStringPointer("ipv6_address: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "link_local_ips",
						Detail:         types.CreateStringPointer("array"),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "mac_address",
						Detail:         types.CreateStringPointer("string"),
						InsertText:     types.CreateStringPointer("mac_address: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "priority",
						Detail:         types.CreateStringPointer("number"),
						InsertText:     types.CreateStringPointer("priority: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
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
					{
						Label:          "bind",
						Detail:         types.CreateStringPointer("object"),
						InsertText:     types.CreateStringPointer("bind:\n          "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "consistency",
						Detail:         types.CreateStringPointer("string"),
						InsertText:     types.CreateStringPointer("consistency: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "image",
						Detail:         types.CreateStringPointer("object"),
						InsertText:     types.CreateStringPointer("image:\n          "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "read_only",
						Detail:         types.CreateStringPointer("boolean or string"),
						InsertText:     types.CreateStringPointer("read_only: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "source",
						Detail:         types.CreateStringPointer("string"),
						InsertText:     types.CreateStringPointer("source: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "target",
						Detail:         types.CreateStringPointer("string"),
						InsertText:     types.CreateStringPointer("target: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "tmpfs",
						Detail:         types.CreateStringPointer("object"),
						InsertText:     types.CreateStringPointer("tmpfs:\n          "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:            "type",
						Detail:           types.CreateStringPointer("string"),
						InsertText:       types.CreateStringPointer("type: ${1|bind,cluster,image,npipe,tmpfs,volume|}"),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "volume",
						Detail:         types.CreateStringPointer("object"),
						InsertText:     types.CreateStringPointer("volume:\n          "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
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
					{
						Label:          "create_host_path",
						Detail:         types.CreateStringPointer("boolean or string"),
						InsertText:     types.CreateStringPointer("create_host_path: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "propagation",
						Detail:         types.CreateStringPointer("string"),
						InsertText:     types.CreateStringPointer("propagation: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:            "recursive",
						Detail:           types.CreateStringPointer("string"),
						InsertText:       types.CreateStringPointer("recursive: ${1|disabled,enabled,readonly,writable|}"),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:            "selinux",
						Detail:           types.CreateStringPointer("string"),
						InsertText:       types.CreateStringPointer("selinux: ${1|Z,z|}"),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
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
					{
						Label:          "labels",
						Detail:         types.CreateStringPointer("array or object"),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "nocopy",
						Detail:         types.CreateStringPointer("boolean or string"),
						InsertText:     types.CreateStringPointer("nocopy: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
					{
						Label:          "subpath",
						Detail:         types.CreateStringPointer("string"),
						InsertText:     types.CreateStringPointer("subpath: "),
						InsertTextMode: types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
					},
				},
			},
		},
		{
			name: "enum properties for a string attribute directly as an item is suggested",
			content: `
services:
  test:
    volumes:
      - type: `,
			line:      4,
			character: 14,
			list: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{Label: "bind", Detail: types.CreateStringPointer("string")},
					{Label: "cluster", Detail: types.CreateStringPointer("string")},
					{Label: "image", Detail: types.CreateStringPointer("string")},
					{Label: "npipe", Detail: types.CreateStringPointer("string")},
					{Label: "tmpfs", Detail: types.CreateStringPointer("string")},
					{Label: "volume", Detail: types.CreateStringPointer("string")},
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
