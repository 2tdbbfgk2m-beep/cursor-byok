package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestManifestForPersonalArm64Release(t *testing.T) {
	dir := t.TempDir()
	for name, data := range map[string]string{
		"config.yml": "info:\n  version: 1.0.1-personal.10\n",
		"notes.md":   "Provider names and inherited model prefixes.",
		"cursor-byok-1.0.1-personal.10-macos-arm64.tar.gz": "abc",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	args := []string{"run", ".", "manifest", "-config", filepath.Join(dir, "config.yml"), "-assets-dir", dir, "-out", filepath.Join(dir, "update.json"), "-repo", "2tdbbfgk2m-beep/cursor-byok", "-notes", filepath.Join(dir, "notes.md"), "-platform", "macos-arm64"}
	if out, err := exec.Command("go", args...).CombinedOutput(); err != nil {
		t.Fatalf("single-platform manifest failed: %v\n%s", err, out)
	}
	data, err := os.ReadFile(filepath.Join(dir, "update.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest updateManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	asset := manifest.Platforms["macos-arm64"]
	if len(manifest.Platforms) != 1 || manifest.Version != "1.0.1-personal.10" || asset.Size != 3 || asset.Checksum != "sha256:ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" || asset.URL != "https://github.com/2tdbbfgk2m-beep/cursor-byok/releases/download/v1.0.1-personal.10/cursor-byok-1.0.1-personal.10-macos-arm64.tar.gz" {
		t.Fatalf("incorrect personal update manifest: %+v", manifest)
	}
	args[len(args)-1] = "unsupported"
	if out, err := exec.Command("go", args...).CombinedOutput(); err == nil {
		t.Fatalf("unsupported platform accepted: %s", out)
	}
}
