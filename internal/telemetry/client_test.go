package telemetry

import (
	"fmt"
	"testing"

	"github.com/docker/docker-language-server/internal/configuration"
	"github.com/stretchr/testify/require"
)

func TestAllow(t *testing.T) {
	eventSettings := []string{"all", "error", "off", ""}
	for i := range eventSettings {
		for j := range eventSettings {
			t.Run(fmt.Sprintf("from %v to %v", eventSettings[i], eventSettings[j]), func(t *testing.T) {
				client := TelemetryClientImpl{}
				client.UpdateTelemetrySetting(eventSettings[i])

				if eventSettings[i] == string(configuration.TelemetrySettingAll) {
					require.True(t, client.allow(true))
					require.True(t, client.allow(false))
				} else if eventSettings[i] == string(configuration.TelemetrySettingError) {
					require.True(t, client.allow(true))
					require.False(t, client.allow(false))
				} else {
					require.False(t, client.allow(true))
					require.False(t, client.allow(false))
				}

				client.UpdateTelemetrySetting(eventSettings[j])

				if eventSettings[j] == string(configuration.TelemetrySettingAll) {
					require.True(t, client.allow(true))
					require.True(t, client.allow(false))
				} else if eventSettings[j] == string(configuration.TelemetrySettingError) {
					require.True(t, client.allow(true))
					require.False(t, client.allow(false))
				} else {
					require.False(t, client.allow(true))
					require.False(t, client.allow(false))
				}
			})
		}
	}
}
