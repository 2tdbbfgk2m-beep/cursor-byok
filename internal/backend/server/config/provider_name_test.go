package config

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// A provider rename must persist and reach Cursor without changing channel identity.
func TestProviderRenamePersistsWithoutChangingModelIdentity(t *testing.T) {
	ctx := context.Background()
	store := NewStore(filepath.Join(t.TempDir(), "config.yaml"), t.TempDir())
	adapter := testModelAdapter("Sonnet", 1)
	cfg, err := store.Save(ctx, Config{ModelAdapters: []ModelAdapterConfig{adapter}})
	if err != nil {
		t.Fatal(err)
	}
	oldID := cfg.ModelAdapters[0].ID
	data, _ := yaml.Marshal(cfg)
	data = []byte(strings.Replace(string(data), "displayName: Sonnet", "displayName: Sonnet\n      providerName: VC", 1))
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatal(err)
	}
	// A later model added to this provider must inherit its name.
	cfg.ModelAdapters = append(cfg.ModelAdapters, testModelAdapter("Opus", 2))
	if _, err := store.Save(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	loaded, err := store.Load(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.ModelAdapters[0].ID != oldID {
		t.Fatal("provider rename changed channel ID")
	}
	data, _ = yaml.Marshal(loaded)
	if strings.Count(string(data), "providerName: VC") != 2 {
		t.Fatalf("provider name was lost or not inherited: %s", data)
	}
	channel, err := resolveModelAdapterChannel(loaded.ModelAdapters, oldID)
	if err != nil {
		t.Fatal(err)
	}
	if channel.Name != "VC / Sonnet" || channel.Model != "Sonnet" {
		t.Fatalf("wrong display/upstream model: %q / %q", channel.Name, channel.Model)
	}
	// Metadata must survive the path used to publish models to Cursor.
	manager, err := NewManager(ctx, store)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := manager.LegacyRuntimeSnapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := yaml.Marshal(snapshot.ModelAdapters)
	if !strings.Contains(string(encoded), "VC") {
		t.Fatal("runtime snapshot lost provider name")
	}
	for _, name := range []string{"Renamed", ""} {
		for i := range loaded.ModelAdapters {
			loaded.ModelAdapters[i].ProviderName = name
		}
		loaded, err = manager.Save(ctx, loaded)
		if err != nil {
			t.Fatal(err)
		}
		channel, err = manager.SelectChannelForModel(ctx, oldID)
		if err != nil {
			t.Fatal(err)
		}
		want := "Renamed / Sonnet"
		if name == "" {
			want = "api.example.com / Sonnet"
		}
		if channel.ID != oldID || channel.Name != want || channel.Model != "Sonnet" {
			t.Fatalf("provider rename/reset broke existing model selection: %+v", channel)
		}
	}
}
