package upstream

import (
	"encoding/json"
	"testing"

	legacyruntime "cursor/internal/runtime"
)

func TestSameNamedModelsDisplayProviderPrefixes(t *testing.T) {
	var adapters []legacyruntime.ModelAdapterConfig
	err := json.Unmarshal([]byte(`[
	 {"id":"channel-a","providerName":"VC","displayName":"Sonnet","modelID":"claude-sonnet","baseURL":"https://a.example"},
	 {"id":"channel-b","providerName":"Fable","displayName":"Sonnet","modelID":"claude-sonnet","baseURL":"https://b.example"},
	 {"id":"channel-c","displayName":"Opus","modelID":"claude-opus","baseURL":"https://legacy.example/v1"}
	]`), &adapters)
	if err != nil {
		t.Fatal(err)
	}
	entries := buildAvailableModelEntries(adapters)
	cli := buildCLIModelDetails(adapters)
	for i, want := range []string{"VC / Sonnet", "Fable / Sonnet", "legacy.example / Opus"} {
		if entries[i]["clientDisplayName"] != want || entries[i]["inputboxShortModelName"] != want || cli[i]["displayName"] != want {
			t.Fatalf("model %d lost provider prefix: IDE=%v CLI=%v, want %q", i, entries[i]["clientDisplayName"], cli[i]["displayName"], want)
		}
		if entries[i]["name"] != adapters[i].ID || cli[i]["modelId"] != adapters[i].ID {
			t.Fatal("prefix changed routing identity")
		}
	}
}
