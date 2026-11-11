package updater

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

type manifestTransport func(*http.Request) (*http.Response, error)

func (transport manifestTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return transport(req)
}

// Catch an upstream merge silently restoring the official update source.
func TestUpdateCheckUsesPersonalReleaseSource(t *testing.T) {
	platform, err := currentPlatformKey()
	if err != nil {
		t.Fatal(err)
	}
	const assetURL = "https://github.com/2tdbbfgk2m-beep/cursor-byok/releases/download/v1.0.1-personal.10/cursor-byok-1.0.1-personal.10-macos-arm64.tar.gz"
	manager := &Manager{client: &http.Client{Transport: manifestTransport(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != "https://github.com/2tdbbfgk2m-beep/cursor-byok/releases/latest/download/update.json" {
			t.Errorf("update check escaped the personal channel: %s", req.URL)
		}
		body := fmt.Sprintf(`{"version":"1.0.1-personal.10","platforms":{%q:{"url":%q,"size":3,"checksum":"sha256:ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"}}}`, platform, assetURL)
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
	})}}
	info, err := manager.fetchUpdateInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if info == nil || info.Version != "1.0.1-personal.10" || info.Asset.URL != assetURL || info.Asset.Size != 3 {
		t.Fatalf("personal update manifest was not selected: %+v", info)
	}
}
