package configuration

import (
	"sync"

	"github.com/docker/docker-language-server/internal/tliron/glsp/protocol"
)

const ConfigTelemetry = "docker.lsp.telemetry"
const ConfigExperimentalVulnerabilityScanning = "docker.lsp.experimental.vulnerabilityScanning"

type TelemetrySetting string

const (
	TelemetrySettingOff   TelemetrySetting = "off"
	TelemetrySettingError TelemetrySetting = "error"
	TelemetrySettingAll   TelemetrySetting = "all"
)

type Configuration struct {
	// docker.lsp.telemetry
	Telemetry    TelemetrySetting `json:"telemetry,omitempty"`
	Experimental Experimental     `json:"experimental"`
}

type Experimental struct {
	// docker.lsp.experimental.vulnerabilityScanning
	VulnerabilityScanning bool `json:"vulnerabilityScanning"`
	// docker.lsp.experimental.scout
	Scout Scout `json:"scout"`
}

type Scout struct {
	CriticalHighVulnerabilities bool `json:"criticalHighVulnerabilities"`
	NotPinnedDigest             bool `json:"notPinnedDigest"`
	RecommendedTag              bool `json:"recommendedTag"`
	Vulnerabilites              bool `json:"vulnerabilites"`
}

var configurations = make(map[protocol.DocumentUri]Configuration)
var lock = sync.RWMutex{}
var defaultConfiguration = Configuration{
	Telemetry: TelemetrySettingAll,
	Experimental: Experimental{
		VulnerabilityScanning: true,
		Scout: Scout{
			CriticalHighVulnerabilities: true,
			NotPinnedDigest:             false,
			RecommendedTag:              true,
			Vulnerabilites:              true,
		},
	},
}

func Documents() []protocol.DocumentUri {
	lock.RLock()
	defer lock.RUnlock()

	documents := []protocol.DocumentUri{}
	for document := range configurations {
		documents = append(documents, document)
	}
	return documents
}

func Get(document protocol.DocumentUri) Configuration {
	lock.RLock()
	defer lock.RUnlock()

	if config, ok := configurations[document]; ok {
		return config
	}
	return defaultConfiguration
}

func Store(document protocol.DocumentUri, configuration Configuration) {
	lock.Lock()
	defer lock.Unlock()
	configurations[document] = configuration
}

func Remove(document protocol.DocumentUri) {
	lock.Lock()
	defer lock.Unlock()
	delete(configurations, document)
}
