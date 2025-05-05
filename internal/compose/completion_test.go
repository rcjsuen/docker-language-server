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

func serviceProperties(line, character, prefixLength protocol.UInteger) []protocol.CompletionItem {
	return []protocol.CompletionItem{
		{
			Label:            "annotations",
			Detail:           types.CreateStringPointer("array or object"),
			TextEdit:         textEdit("annotations:\n      ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "attach",
			Detail:           types.CreateStringPointer("boolean or string"),
			TextEdit:         textEdit("attach: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "blkio_config",
			Detail:           types.CreateStringPointer("object"),
			TextEdit:         textEdit("blkio_config:\n      ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "build",
			Detail:           types.CreateStringPointer("object or string"),
			TextEdit:         textEdit("build:", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "cap_add",
			Detail:           types.CreateStringPointer("array"),
			TextEdit:         textEdit("cap_add:\n      - ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "cap_drop",
			Detail:           types.CreateStringPointer("array"),
			TextEdit:         textEdit("cap_drop:\n      - ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "cgroup",
			Detail:           types.CreateStringPointer("string"),
			TextEdit:         textEdit("cgroup: ${1|host,private|}", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "cgroup_parent",
			Detail:           types.CreateStringPointer("string"),
			TextEdit:         textEdit("cgroup_parent: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "command",
			Detail:           types.CreateStringPointer("array or null or string"),
			TextEdit:         textEdit("command:", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "configs",
			Detail:           types.CreateStringPointer("array"),
			TextEdit:         textEdit("configs:\n      - ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "container_name",
			Detail:           types.CreateStringPointer("string"),
			TextEdit:         textEdit("container_name: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "cpu_count",
			Detail:           types.CreateStringPointer("integer or string"),
			TextEdit:         textEdit("cpu_count: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "cpu_percent",
			Detail:           types.CreateStringPointer("integer or string"),
			TextEdit:         textEdit("cpu_percent: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "cpu_period",
			Detail:           types.CreateStringPointer("number or string"),
			TextEdit:         textEdit("cpu_period: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "cpu_quota",
			Detail:           types.CreateStringPointer("number or string"),
			TextEdit:         textEdit("cpu_quota: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "cpu_rt_period",
			Detail:           types.CreateStringPointer("number or string"),
			TextEdit:         textEdit("cpu_rt_period: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "cpu_rt_runtime",
			Detail:           types.CreateStringPointer("number or string"),
			TextEdit:         textEdit("cpu_rt_runtime: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "cpu_shares",
			Detail:           types.CreateStringPointer("number or string"),
			TextEdit:         textEdit("cpu_shares: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "cpus",
			Detail:           types.CreateStringPointer("number or string"),
			TextEdit:         textEdit("cpus: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "cpuset",
			Detail:           types.CreateStringPointer("string"),
			TextEdit:         textEdit("cpuset: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "credential_spec",
			Detail:           types.CreateStringPointer("object"),
			TextEdit:         textEdit("credential_spec:\n      ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "depends_on",
			Detail:           types.CreateStringPointer("array or object"),
			TextEdit:         textEdit("depends_on:\n      ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "deploy",
			Detail:           types.CreateStringPointer("null or object"),
			TextEdit:         textEdit("deploy:", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "develop",
			Detail:           types.CreateStringPointer("null or object"),
			TextEdit:         textEdit("develop:", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "device_cgroup_rules",
			Detail:           types.CreateStringPointer("array"),
			TextEdit:         textEdit("device_cgroup_rules:\n      - ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "devices",
			Detail:           types.CreateStringPointer("array"),
			TextEdit:         textEdit("devices:\n      - ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "dns",
			Detail:           types.CreateStringPointer("array or string"),
			TextEdit:         textEdit("dns:", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "dns_opt",
			Detail:           types.CreateStringPointer("array"),
			TextEdit:         textEdit("dns_opt:\n      - ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "dns_search",
			Detail:           types.CreateStringPointer("array or string"),
			TextEdit:         textEdit("dns_search:", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "domainname",
			Detail:           types.CreateStringPointer("string"),
			TextEdit:         textEdit("domainname: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "entrypoint",
			Detail:           types.CreateStringPointer("array or null or string"),
			TextEdit:         textEdit("entrypoint:", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "env_file",
			Detail:           types.CreateStringPointer("array or string"),
			TextEdit:         textEdit("env_file:", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "environment",
			Detail:           types.CreateStringPointer("array or object"),
			TextEdit:         textEdit("environment:\n      ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "expose",
			Detail:           types.CreateStringPointer("array"),
			TextEdit:         textEdit("expose:\n      - ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "extends",
			Detail:           types.CreateStringPointer("object or string"),
			TextEdit:         textEdit("extends:", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "external_links",
			Detail:           types.CreateStringPointer("array"),
			TextEdit:         textEdit("external_links:\n      - ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "extra_hosts",
			Detail:           types.CreateStringPointer("array or object"),
			TextEdit:         textEdit("extra_hosts:\n      ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "gpus",
			Detail:           types.CreateStringPointer("array or string"),
			TextEdit:         textEdit("gpus:", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "group_add",
			Detail:           types.CreateStringPointer("array"),
			TextEdit:         textEdit("group_add:\n      - ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "healthcheck",
			Detail:           types.CreateStringPointer("object"),
			TextEdit:         textEdit("healthcheck:\n      ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "hostname",
			Detail:           types.CreateStringPointer("string"),
			TextEdit:         textEdit("hostname: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "image",
			Detail:           types.CreateStringPointer("string"),
			TextEdit:         textEdit("image: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "init",
			Detail:           types.CreateStringPointer("boolean or string"),
			TextEdit:         textEdit("init: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "ipc",
			Detail:           types.CreateStringPointer("string"),
			TextEdit:         textEdit("ipc: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "isolation",
			Detail:           types.CreateStringPointer("string"),
			TextEdit:         textEdit("isolation: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "label_file",
			Detail:           types.CreateStringPointer("array or string"),
			TextEdit:         textEdit("label_file:", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "labels",
			Detail:           types.CreateStringPointer("array or object"),
			TextEdit:         textEdit("labels:\n      ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "links",
			Detail:           types.CreateStringPointer("array"),
			TextEdit:         textEdit("links:\n      - ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "logging",
			Detail:           types.CreateStringPointer("object"),
			TextEdit:         textEdit("logging:\n      ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "mac_address",
			Detail:           types.CreateStringPointer("string"),
			TextEdit:         textEdit("mac_address: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "mem_limit",
			Detail:           types.CreateStringPointer("number or string"),
			TextEdit:         textEdit("mem_limit: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "mem_reservation",
			Detail:           types.CreateStringPointer("integer or string"),
			TextEdit:         textEdit("mem_reservation: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "mem_swappiness",
			Detail:           types.CreateStringPointer("integer or string"),
			TextEdit:         textEdit("mem_swappiness: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "memswap_limit",
			Detail:           types.CreateStringPointer("number or string"),
			TextEdit:         textEdit("memswap_limit: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "network_mode",
			Detail:           types.CreateStringPointer("string"),
			TextEdit:         textEdit("network_mode: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "networks",
			Detail:           types.CreateStringPointer("array or object"),
			TextEdit:         textEdit("networks:\n      ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "oom_kill_disable",
			Detail:           types.CreateStringPointer("boolean or string"),
			TextEdit:         textEdit("oom_kill_disable: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "oom_score_adj",
			Detail:           types.CreateStringPointer("integer or string"),
			TextEdit:         textEdit("oom_score_adj: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "pid",
			Detail:           types.CreateStringPointer("null or string"),
			TextEdit:         textEdit("pid: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "pids_limit",
			Detail:           types.CreateStringPointer("number or string"),
			TextEdit:         textEdit("pids_limit: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "platform",
			Detail:           types.CreateStringPointer("string"),
			TextEdit:         textEdit("platform: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "ports",
			Detail:           types.CreateStringPointer("array"),
			TextEdit:         textEdit("ports:\n      - ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "post_start",
			Detail:           types.CreateStringPointer("array"),
			TextEdit:         textEdit("post_start:\n      - command:", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "pre_stop",
			Detail:           types.CreateStringPointer("array"),
			TextEdit:         textEdit("pre_stop:\n      - command:", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "privileged",
			Detail:           types.CreateStringPointer("boolean or string"),
			TextEdit:         textEdit("privileged: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "profiles",
			Detail:           types.CreateStringPointer("array"),
			TextEdit:         textEdit("profiles:\n      - ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "provider",
			Detail:           types.CreateStringPointer("object"),
			TextEdit:         textEdit("provider:\n      ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "pull_policy",
			Detail:           types.CreateStringPointer("string"),
			TextEdit:         textEdit("pull_policy: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "pull_refresh_after",
			Detail:           types.CreateStringPointer("string"),
			TextEdit:         textEdit("pull_refresh_after: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "read_only",
			Detail:           types.CreateStringPointer("boolean or string"),
			TextEdit:         textEdit("read_only: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "restart",
			Detail:           types.CreateStringPointer("string"),
			TextEdit:         textEdit("restart: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "runtime",
			Detail:           types.CreateStringPointer("string"),
			TextEdit:         textEdit("runtime: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "scale",
			Detail:           types.CreateStringPointer("integer or string"),
			TextEdit:         textEdit("scale: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "secrets",
			Detail:           types.CreateStringPointer("array"),
			TextEdit:         textEdit("secrets:\n      - ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "security_opt",
			Detail:           types.CreateStringPointer("array"),
			TextEdit:         textEdit("security_opt:\n      - ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "shm_size",
			Detail:           types.CreateStringPointer("number or string"),
			TextEdit:         textEdit("shm_size: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "stdin_open",
			Detail:           types.CreateStringPointer("boolean or string"),
			TextEdit:         textEdit("stdin_open: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "stop_grace_period",
			Detail:           types.CreateStringPointer("string"),
			TextEdit:         textEdit("stop_grace_period: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "stop_signal",
			Detail:           types.CreateStringPointer("string"),
			TextEdit:         textEdit("stop_signal: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "storage_opt",
			Detail:           types.CreateStringPointer("object"),
			TextEdit:         textEdit("storage_opt:\n      ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "sysctls",
			Detail:           types.CreateStringPointer("array or object"),
			TextEdit:         textEdit("sysctls:\n      ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "tmpfs",
			Detail:           types.CreateStringPointer("array or string"),
			TextEdit:         textEdit("tmpfs:", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "tty",
			Detail:           types.CreateStringPointer("boolean or string"),
			TextEdit:         textEdit("tty: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "ulimits",
			Detail:           types.CreateStringPointer("object"),
			TextEdit:         textEdit("ulimits:\n      ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "user",
			Detail:           types.CreateStringPointer("string"),
			TextEdit:         textEdit("user: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "userns_mode",
			Detail:           types.CreateStringPointer("string"),
			TextEdit:         textEdit("userns_mode: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "uts",
			Detail:           types.CreateStringPointer("string"),
			TextEdit:         textEdit("uts: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "volumes",
			Detail:           types.CreateStringPointer("array"),
			TextEdit:         textEdit("volumes:\n      - ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "volumes_from",
			Detail:           types.CreateStringPointer("array"),
			TextEdit:         textEdit("volumes_from:\n      - ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
		{
			Label:            "working_dir",
			Detail:           types.CreateStringPointer("string"),
			TextEdit:         textEdit("working_dir: ", line, character, prefixLength),
			InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
			InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
		},
	}
}

func TestCompletion_Schema(t *testing.T) {
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
					{
						Label:         "configs",
						Documentation: "Configurations for services in the project",
					},
					{
						Label:         "include",
						Documentation: "compose sub-projects to be included.",
					},
					{
						Label:         "name",
						Documentation: "define the Compose project name, until user defines one explicitly.",
					},
					{
						Label:         "networks",
						Documentation: "Networks that are shared among multiple services",
					},
					{
						Label:         "secrets",
						Documentation: "Secrets that are shared among multiple services",
					},
					{
						Label:         "services",
						Documentation: "The services in your project",
					},
					{
						Label:         "version",
						Documentation: "declared for backward compatibility, ignored.",
					},
					{
						Label:         "volumes",
						Documentation: "Named volumes that are shared among multiple services",
					},
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
						Label:            "content",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("content: ", 3, 4, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "environment",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("environment: ", 3, 4, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "external",
						Detail:           types.CreateStringPointer("boolean or object or string"),
						TextEdit:         textEdit("external:", 3, 4, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "file",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("file: ", 3, 4, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "labels",
						Detail:           types.CreateStringPointer("array or object"),
						TextEdit:         textEdit("labels:\n      ", 3, 4, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "name",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("name: ", 3, 4, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "template_driver",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("template_driver: ", 3, 4, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
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
						Label:            "attachable",
						Detail:           types.CreateStringPointer("boolean or string"),
						TextEdit:         textEdit("attachable: ", 3, 4, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "driver",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("driver: ", 3, 4, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "driver_opts",
						Detail:           types.CreateStringPointer("object"),
						TextEdit:         textEdit("driver_opts:\n      ", 3, 4, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "enable_ipv4",
						Detail:           types.CreateStringPointer("boolean or string"),
						TextEdit:         textEdit("enable_ipv4: ", 3, 4, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "enable_ipv6",
						Detail:           types.CreateStringPointer("boolean or string"),
						TextEdit:         textEdit("enable_ipv6: ", 3, 4, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "external",
						Detail:           types.CreateStringPointer("boolean or object or string"),
						TextEdit:         textEdit("external:", 3, 4, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "internal",
						Detail:           types.CreateStringPointer("boolean or string"),
						TextEdit:         textEdit("internal: ", 3, 4, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "ipam",
						Detail:           types.CreateStringPointer("object"),
						TextEdit:         textEdit("ipam:\n      ", 3, 4, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "labels",
						Detail:           types.CreateStringPointer("array or object"),
						TextEdit:         textEdit("labels:\n      ", 3, 4, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "name",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("name: ", 3, 4, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
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
						Label:            "driver",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("driver: ", 3, 4, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "driver_opts",
						Detail:           types.CreateStringPointer("object"),
						TextEdit:         textEdit("driver_opts:\n      ", 3, 4, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "environment",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("environment: ", 3, 4, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "external",
						Detail:           types.CreateStringPointer("boolean or object or string"),
						TextEdit:         textEdit("external:", 3, 4, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "file",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("file: ", 3, 4, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "labels",
						Detail:           types.CreateStringPointer("array or object"),
						TextEdit:         textEdit("labels:\n      ", 3, 4, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "name",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("name: ", 3, 4, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "template_driver",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("template_driver: ", 3, 4, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
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
				Items: serviceProperties(3, 4, 0),
			},
		},
		{
			name: "prefix of service attributes",
			content: `
services:
  test:
    a`,
			line:      3,
			character: 5,
			list: &protocol.CompletionList{
				Items: serviceProperties(3, 5, 1),
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
						Label:            "driver",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("driver: ", 3, 4, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "driver_opts",
						Detail:           types.CreateStringPointer("object"),
						TextEdit:         textEdit("driver_opts:\n      ", 3, 4, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "external",
						Detail:           types.CreateStringPointer("boolean or object or string"),
						TextEdit:         textEdit("external:", 3, 4, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "labels",
						Detail:           types.CreateStringPointer("array or object"),
						TextEdit:         textEdit("labels:\n      ", 3, 4, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "name",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("name: ", 3, 4, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
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
						Label:            "device_read_bps",
						Detail:           types.CreateStringPointer("array"),
						TextEdit:         textEdit("device_read_bps:\n        - ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "device_read_iops",
						Detail:           types.CreateStringPointer("array"),
						TextEdit:         textEdit("device_read_iops:\n        - ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "device_write_bps",
						Detail:           types.CreateStringPointer("array"),
						TextEdit:         textEdit("device_write_bps:\n        - ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "device_write_iops",
						Detail:           types.CreateStringPointer("array"),
						TextEdit:         textEdit("device_write_iops:\n        - ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "weight",
						Detail:           types.CreateStringPointer("integer or string"),
						TextEdit:         textEdit("weight: ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "weight_device",
						Detail:           types.CreateStringPointer("array"),
						TextEdit:         textEdit("weight_device:\n        - ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
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
						Label:            "endpoint_mode",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("endpoint_mode: ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "labels",
						Detail:           types.CreateStringPointer("array or object"),
						TextEdit:         textEdit("labels:\n        ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "mode",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("mode: ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "placement",
						Detail:           types.CreateStringPointer("object"),
						TextEdit:         textEdit("placement:\n        ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "replicas",
						Detail:           types.CreateStringPointer("integer or string"),
						TextEdit:         textEdit("replicas: ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "resources",
						Detail:           types.CreateStringPointer("object"),
						TextEdit:         textEdit("resources:\n        ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "restart_policy",
						Detail:           types.CreateStringPointer("object"),
						TextEdit:         textEdit("restart_policy:\n        ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "rollback_config",
						Detail:           types.CreateStringPointer("object"),
						TextEdit:         textEdit("rollback_config:\n        ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "update_config",
						Detail:           types.CreateStringPointer("object"),
						TextEdit:         textEdit("update_config:\n        ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
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
						Label:            "limits",
						Detail:           types.CreateStringPointer("object"),
						TextEdit:         textEdit("limits:\n          ", 5, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "reservations",
						Detail:           types.CreateStringPointer("object"),
						TextEdit:         textEdit("reservations:\n          ", 5, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
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
						Label:            "watch",
						Detail:           types.CreateStringPointer("array"),
						TextEdit:         textEdit("watch:\n        - action: ${1}\n          path: ${2}", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
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
						Label:            "additional_contexts",
						Detail:           types.CreateStringPointer("array or object"),
						TextEdit:         textEdit("additional_contexts:\n        ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "args",
						Detail:           types.CreateStringPointer("array or object"),
						TextEdit:         textEdit("args:\n        ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "cache_from",
						Detail:           types.CreateStringPointer("array"),
						TextEdit:         textEdit("cache_from:\n        - ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "cache_to",
						Detail:           types.CreateStringPointer("array"),
						TextEdit:         textEdit("cache_to:\n        - ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "context",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("context: ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "dockerfile",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("dockerfile: ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "dockerfile_inline",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("dockerfile_inline: ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "entitlements",
						Detail:           types.CreateStringPointer("array"),
						TextEdit:         textEdit("entitlements:\n        - ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "extra_hosts",
						Detail:           types.CreateStringPointer("array or object"),
						TextEdit:         textEdit("extra_hosts:\n        ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "isolation",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("isolation: ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "labels",
						Detail:           types.CreateStringPointer("array or object"),
						TextEdit:         textEdit("labels:\n        ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "network",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("network: ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "no_cache",
						Detail:           types.CreateStringPointer("boolean or string"),
						TextEdit:         textEdit("no_cache: ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "platforms",
						Detail:           types.CreateStringPointer("array"),
						TextEdit:         textEdit("platforms:\n        - ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "privileged",
						Detail:           types.CreateStringPointer("boolean or string"),
						TextEdit:         textEdit("privileged: ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "pull",
						Detail:           types.CreateStringPointer("boolean or string"),
						TextEdit:         textEdit("pull: ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "secrets",
						Detail:           types.CreateStringPointer("array"),
						TextEdit:         textEdit("secrets:\n        - ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "shm_size",
						Detail:           types.CreateStringPointer("integer or string"),
						TextEdit:         textEdit("shm_size: ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "ssh",
						Detail:           types.CreateStringPointer("array or object"),
						TextEdit:         textEdit("ssh:\n        ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "tags",
						Detail:           types.CreateStringPointer("array"),
						TextEdit:         textEdit("tags:\n        - ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "target",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("target: ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "ulimits",
						Detail:           types.CreateStringPointer("object"),
						TextEdit:         textEdit("ulimits:\n        ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
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
						Label:            "device_read_bps",
						Detail:           types.CreateStringPointer("array"),
						TextEdit:         textEdit("device_read_bps:\n        - ", 5, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "device_read_iops",
						Detail:           types.CreateStringPointer("array"),
						TextEdit:         textEdit("device_read_iops:\n        - ", 5, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "device_write_bps",
						Detail:           types.CreateStringPointer("array"),
						TextEdit:         textEdit("device_write_bps:\n        - ", 5, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "device_write_iops",
						Detail:           types.CreateStringPointer("array"),
						TextEdit:         textEdit("device_write_iops:\n        - ", 5, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "weight",
						Detail:           types.CreateStringPointer("integer or string"),
						TextEdit:         textEdit("weight: ", 5, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "weight_device",
						Detail:           types.CreateStringPointer("array"),
						TextEdit:         textEdit("weight_device:\n        - ", 5, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
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
						Label:            "gid",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("gid: ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "mode",
						Detail:           types.CreateStringPointer("number or string"),
						TextEdit:         textEdit("mode: ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "source",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("source: ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "target",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("target: ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "uid",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("uid: ", 4, 6, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
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
						Label:            "hard",
						Detail:           types.CreateStringPointer("integer or string"),
						TextEdit:         textEdit("hard: ", 6, 10, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "soft",
						Detail:           types.CreateStringPointer("integer or string"),
						TextEdit:         textEdit("soft: ", 6, 10, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
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
				Items: serviceProperties(4, 4, 0),
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
				Items: serviceProperties(5, 4, 0),
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
						Label:            "aliases",
						Detail:           types.CreateStringPointer("array"),
						TextEdit:         textEdit("aliases:\n          - ", 5, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "driver_opts",
						Detail:           types.CreateStringPointer("object"),
						TextEdit:         textEdit("driver_opts:\n          ", 5, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "gw_priority",
						Detail:           types.CreateStringPointer("number"),
						TextEdit:         textEdit("gw_priority: ", 5, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "interface_name",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("interface_name: ", 5, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "ipv4_address",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("ipv4_address: ", 5, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "ipv6_address",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("ipv6_address: ", 5, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "link_local_ips",
						Detail:           types.CreateStringPointer("array"),
						TextEdit:         textEdit("link_local_ips:\n          - ", 5, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "mac_address",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("mac_address: ", 5, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "priority",
						Detail:           types.CreateStringPointer("number"),
						TextEdit:         textEdit("priority: ", 5, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
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
				Items: serviceProperties(5, 4, 0),
			},
		},
		{
			name: "properties of a volumes array item with no content",
			content: `
services:
  test:
    image: alpine
    volumes:
      - `,
			line:      5,
			character: 8,
			list: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{
						Label:            "bind",
						Detail:           types.CreateStringPointer("object"),
						TextEdit:         textEdit("bind:\n          ", 5, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "consistency",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("consistency: ", 5, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "image",
						Detail:           types.CreateStringPointer("object"),
						TextEdit:         textEdit("image:\n          ", 5, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "read_only",
						Detail:           types.CreateStringPointer("boolean or string"),
						TextEdit:         textEdit("read_only: ", 5, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "source",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("source: ", 5, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "target",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("target: ", 5, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "tmpfs",
						Detail:           types.CreateStringPointer("object"),
						TextEdit:         textEdit("tmpfs:\n          ", 5, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "type",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("type: ${1|bind,cluster,image,npipe,tmpfs,volume|}", 5, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "volume",
						Detail:           types.CreateStringPointer("object"),
						TextEdit:         textEdit("volume:\n          ", 5, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
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
					{
						Label:            "bind",
						Detail:           types.CreateStringPointer("object"),
						TextEdit:         textEdit("bind:\n          ", 6, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "consistency",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("consistency: ", 6, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "image",
						Detail:           types.CreateStringPointer("object"),
						TextEdit:         textEdit("image:\n          ", 6, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "read_only",
						Detail:           types.CreateStringPointer("boolean or string"),
						TextEdit:         textEdit("read_only: ", 6, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "source",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("source: ", 6, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "target",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("target: ", 6, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "tmpfs",
						Detail:           types.CreateStringPointer("object"),
						TextEdit:         textEdit("tmpfs:\n          ", 6, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "type",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("type: ${1|bind,cluster,image,npipe,tmpfs,volume|}", 6, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "volume",
						Detail:           types.CreateStringPointer("object"),
						TextEdit:         textEdit("volume:\n          ", 6, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
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
						Label:            "create_host_path",
						Detail:           types.CreateStringPointer("boolean or string"),
						TextEdit:         textEdit("create_host_path: ", 7, 10, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "propagation",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("propagation: ", 7, 10, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "recursive",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("recursive: ${1|disabled,enabled,readonly,writable|}", 7, 10, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "selinux",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("selinux: ${1|Z,z|}", 7, 10, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
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
						Label:            "labels",
						Detail:           types.CreateStringPointer("array or object"),
						TextEdit:         textEdit("labels:\n            ", 7, 10, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "nocopy",
						Detail:           types.CreateStringPointer("boolean or string"),
						TextEdit:         textEdit("nocopy: ", 7, 10, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "subpath",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("subpath: ", 7, 10, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
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
					{
						Label:    "bind",
						Detail:   types.CreateStringPointer("string"),
						TextEdit: textEdit("bind", 4, 14, 0),
					},
					{
						Label:    "cluster",
						Detail:   types.CreateStringPointer("string"),
						TextEdit: textEdit("cluster", 4, 14, 0),
					},
					{
						Label:    "image",
						Detail:   types.CreateStringPointer("string"),
						TextEdit: textEdit("image", 4, 14, 0),
					},
					{
						Label:    "npipe",
						Detail:   types.CreateStringPointer("string"),
						TextEdit: textEdit("npipe", 4, 14, 0),
					},
					{
						Label:    "tmpfs",
						Detail:   types.CreateStringPointer("string"),
						TextEdit: textEdit("tmpfs", 4, 14, 0),
					},
					{
						Label:    "volume",
						Detail:   types.CreateStringPointer("string"),
						TextEdit: textEdit("volume", 4, 14, 0),
					},
				},
			},
		},
		{
			name: "enum properties for a string attribute considers the prefix for the TextEdit",
			content: `
services:
  test:
    volumes:
      - type: b`,
			line:      4,
			character: 15,
			list: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{
						Label:    "bind",
						Detail:   types.CreateStringPointer("string"),
						TextEdit: textEdit("bind", 4, 15, 1),
					},
					{
						Label:    "cluster",
						Detail:   types.CreateStringPointer("string"),
						TextEdit: textEdit("cluster", 4, 15, 1),
					},
					{
						Label:    "image",
						Detail:   types.CreateStringPointer("string"),
						TextEdit: textEdit("image", 4, 15, 1),
					},
					{
						Label:    "npipe",
						Detail:   types.CreateStringPointer("string"),
						TextEdit: textEdit("npipe", 4, 15, 1),
					},
					{
						Label:    "tmpfs",
						Detail:   types.CreateStringPointer("string"),
						TextEdit: textEdit("tmpfs", 4, 15, 1),
					},
					{
						Label:    "volume",
						Detail:   types.CreateStringPointer("string"),
						TextEdit: textEdit("volume", 4, 15, 1),
					},
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
					{
						Label:    "Z",
						Detail:   types.CreateStringPointer("string"),
						TextEdit: textEdit("Z", 6, 19, 0),
					},
					{
						Label:    "z",
						Detail:   types.CreateStringPointer("string"),
						TextEdit: textEdit("z", 6, 19, 0),
					},
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
		{
			name: "unexpected array for networks",
			content: `
networks:
- 
- 
- `,
			line:      4,
			character: 2,
			list:      nil,
		},
		{
			name: "configs should not suggest attributes as-is",
			content: `
services:
  test:
    configs:
      `,
			line:      4,
			character: 6,
			list:      nil,
		},
		{
			name: "array items should still consider indentation when calculating completion items",
			content: `
services:
  test:
    configs:
      - mode: 0
      `,
			line:      5,
			character: 6,
			list:      nil,
		},
		{
			name: "reservations completion",
			content: `
services:
  test:
    image: redis
    deploy:
      resources:
        reservations:
          `,
			line:      7,
			character: 10,
			list: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{
						Label:            "cpus",
						Detail:           types.CreateStringPointer("number or string"),
						TextEdit:         textEdit("cpus: ", 7, 10, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "devices",
						Detail:           types.CreateStringPointer("array"),
						TextEdit:         textEdit("devices:\n            - capabilities:\n              - ", 7, 10, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "generic_resources",
						Detail:           types.CreateStringPointer("array"),
						TextEdit:         textEdit("generic_resources:\n            - ", 7, 10, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "memory",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("memory: ", 7, 10, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
				},
			},
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
			require.Equal(t, tc.list, list)
		})
	}
}

func TestCompletion_Custom(t *testing.T) {
	testCases := []struct {
		name      string
		content   string
		line      uint32
		character uint32
		list      *protocol.CompletionList
	}{
		{
			name: "depends_on array items",
			content: `
services:
  test:
    image: alpine
    depends_on:
      - 
  test2:
    image: alpine`,
			line:      5,
			character: 8,
			list: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{
						Label:    "test2",
						TextEdit: textEdit("test2", 5, 8, 0),
					},
				},
			},
		},
		{
			name: "depends_on array items across two files",
			content: `
---
services:
  test:
    image: alpine
    depends_on:
      - 
---
services:
  test2:
    image: alpine`,
			line:      6,
			character: 8,
			list: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{
						Label:    "test2",
						TextEdit: textEdit("test2", 6, 8, 0),
					},
				},
			},
		},
		{
			name: "depends_on array items with a prefix",
			content: `
services:
  test:
    image: alpine
    depends_on:
      - t
  test2:
    image: alpine`,
			line:      5,
			character: 9,
			list: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{
						Label:    "test2",
						TextEdit: textEdit("test2", 5, 9, 1),
					},
				},
			},
		},
		{
			name: "depends_on service object",
			content: `
services:
  test:
    image: alpine
    depends_on:
      
  test2:
    image: alpine`,
			line:      5,
			character: 6,
			list: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{
						Label:    "test2",
						TextEdit: textEdit("test2", 5, 6, 0),
					},
				},
			},
		},
		{
			name: "networks array items",
			content: `
services:
  test:
    image: alpine
    networks:
      - 
networks:
  test2:
    image: alpine`,
			line:      5,
			character: 8,
			list: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{
						Label:    "test2",
						TextEdit: textEdit("test2", 5, 8, 0),
					},
				},
			},
		},
		{
			name: "networks array items across two files",
			content: `
---
services:
  test:
    image: alpine
    networks:
      - 
---
networks:
  test2:`,
			line:      6,
			character: 8,
			list: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{
						Label:    "test2",
						TextEdit: textEdit("test2", 6, 8, 0),
					},
				},
			},
		},
		{
			name: "networks array items with a prefix",
			content: `
services:
  test:
    image: alpine
    networks:
      - t
networks:
  test2:`,
			line:      5,
			character: 9,
			list: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{
						Label:    "test2",
						TextEdit: textEdit("test2", 5, 9, 1),
					},
				},
			},
		},
		{
			name: "networks service object",
			content: `
services:
  test:
    image: alpine
    networks:
      
networks:
  test2:`,
			line:      5,
			character: 6,
			list: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{
						Label:    "test2",
						TextEdit: textEdit("test2", 5, 6, 0),
					},
				},
			},
		},
		{
			name: "volumes array items",
			content: `
services:
  test:
    image: alpine
    volumes:
      - 
volumes:
  test2:`,
			line:      5,
			character: 8,
			list: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{
						Label:            "test2",
						TextEdit:         textEdit("test2:${1:/container/path}", 5, 8, 0),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
				},
			},
		},
		{
			name: "volumes array items across two files",
			content: `
---
services:
  test:
    image: alpine
    volumes:
      - 
---
volumes:
  test2:`,
			line:      6,
			character: 8,
			list: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{
						Label:            "test2",
						TextEdit:         textEdit("test2:${1:/container/path}", 6, 8, 0),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
				},
			},
		},
		{
			name: "volumes array items with a prefix",
			content: `
services:
  test:
    image: alpine
    volumes:
      - t
volumes:
  test2:`,
			line:      5,
			character: 9,
			list: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{
						Label:            "test2",
						TextEdit:         textEdit("test2:${1:/container/path}", 5, 9, 1),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
				},
			},
		},
		{
			name: "volumes service object",
			content: `
services:
  test:
    image: alpine
    volumes:
      
volumes:
  test2:`,
			line:      5,
			character: 6,
			list: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{
						Label:            "test2",
						TextEdit:         textEdit("test2:${1:/container/path}", 5, 6, 0),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
				},
			},
		},
		{
			name: "configs array items",
			content: `
services:
  test:
    image: alpine
    configs:
      - 
configs:
  test2:
    file: ./httpd.conf`,
			line:      5,
			character: 8,
			list: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{
						Label:            "gid",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("gid: ", 5, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "mode",
						Detail:           types.CreateStringPointer("number or string"),
						TextEdit:         textEdit("mode: ", 5, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "source",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("source: ", 5, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "target",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("target: ", 5, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:    "test2",
						TextEdit: textEdit("test2", 5, 8, 0),
					},
					{
						Label:            "uid",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("uid: ", 5, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
				},
			},
		},
		{
			name: "configs array items across two files",
			content: `
---
services:
  test:
    image: alpine
    configs:
      - 
---
configs:
  test2:
    file: ./httpd.conf`,
			line:      6,
			character: 8,
			list: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{
						Label:            "gid",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("gid: ", 6, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "mode",
						Detail:           types.CreateStringPointer("number or string"),
						TextEdit:         textEdit("mode: ", 6, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "source",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("source: ", 6, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "target",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("target: ", 6, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:    "test2",
						TextEdit: textEdit("test2", 6, 8, 0),
					},
					{
						Label:            "uid",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("uid: ", 6, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
				},
			},
		},
		{
			name: "configs array items with a prefix",
			content: `
services:
  test:
    image: alpine
    configs:
      - t
configs:
  test2:
    file: ./httpd.conf`,
			line:      5,
			character: 9,
			list: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{
						Label:            "gid",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("gid: ", 5, 9, 1),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "mode",
						Detail:           types.CreateStringPointer("number or string"),
						TextEdit:         textEdit("mode: ", 5, 9, 1),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "source",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("source: ", 5, 9, 1),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "target",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("target: ", 5, 9, 1),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:    "test2",
						TextEdit: textEdit("test2", 5, 9, 1),
					},
					{
						Label:            "uid",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("uid: ", 5, 9, 1),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
				},
			},
		},
		{
			name: "secrets array items",
			content: `
services:
  test:
    image: alpine
    secrets:
      - 
secrets:
  test2:
    file: ./httpd.conf`,
			line:      5,
			character: 8,
			list: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{
						Label:            "gid",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("gid: ", 5, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "mode",
						Detail:           types.CreateStringPointer("number or string"),
						TextEdit:         textEdit("mode: ", 5, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "source",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("source: ", 5, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "target",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("target: ", 5, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:    "test2",
						TextEdit: textEdit("test2", 5, 8, 0),
					},
					{
						Label:            "uid",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("uid: ", 5, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
				},
			},
		},
		{
			name: "secrets array items across two files",
			content: `
---
services:
  test:
    image: alpine
    secrets:
      - 
---
secrets:
  test2:
    file: ./httpd.conf`,
			line:      6,
			character: 8,
			list: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{
						Label:            "gid",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("gid: ", 6, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "mode",
						Detail:           types.CreateStringPointer("number or string"),
						TextEdit:         textEdit("mode: ", 6, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "source",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("source: ", 6, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "target",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("target: ", 6, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:    "test2",
						TextEdit: textEdit("test2", 6, 8, 0),
					},
					{
						Label:            "uid",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("uid: ", 6, 8, 0),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
				},
			},
		},
		{
			name: "secrets array items with a prefix",
			content: `
services:
  test:
    image: alpine
    secrets:
      - t
secrets:
  test2:
    file: ./httpd.conf`,
			line:      5,
			character: 9,
			list: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{
						Label:            "gid",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("gid: ", 5, 9, 1),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "mode",
						Detail:           types.CreateStringPointer("number or string"),
						TextEdit:         textEdit("mode: ", 5, 9, 1),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "source",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("source: ", 5, 9, 1),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:            "target",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("target: ", 5, 9, 1),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
					{
						Label:    "test2",
						TextEdit: textEdit("test2", 5, 9, 1),
					},
					{
						Label:            "uid",
						Detail:           types.CreateStringPointer("string"),
						TextEdit:         textEdit("uid: ", 5, 9, 1),
						InsertTextMode:   types.CreateInsertTextModePointer(protocol.InsertTextModeAsIs),
						InsertTextFormat: types.CreateInsertTextFormatPointer(protocol.InsertTextFormatSnippet),
					},
				},
			},
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

func textEdit(newText string, line, character, prefixLength protocol.UInteger) protocol.TextEdit {
	return protocol.TextEdit{
		NewText: newText,
		Range: protocol.Range{
			Start: protocol.Position{
				Line:      line,
				Character: character - prefixLength,
			},
			End: protocol.Position{
				Line:      line,
				Character: character,
			},
		},
	}
}
