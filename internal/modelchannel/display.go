package modelchannel

import (
	"net/url"
	"strings"
)

// DisplayName adds provider metadata without changing either routing or upstream model IDs.
func DisplayName(providerName, baseURL, displayName, modelID string) string {
	name := strings.TrimSpace(displayName)
	if name == "" {
		name = strings.TrimSpace(modelID)
	}
	provider := strings.TrimSpace(providerName)
	if provider == "" {
		if parsed, err := url.Parse(strings.TrimSpace(baseURL)); err == nil {
			provider = parsed.Hostname()
		}
	}
	if provider == "" {
		return name
	}
	return provider + " / " + name
}
