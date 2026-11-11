package updater

import "testing"

func TestPersonalReleaseVersionOrdering(t *testing.T) {
	for _, tc := range []struct{ newer, older string }{
		{"1.0.1-personal.10", "1.0.1-personal.9"},
		{"1.0.1-personal.11", "1.0.1-personal.10"},
		{"1.0.2-personal.1", "1.0.1-personal.99"},
		{"1.0.1", "1.0.1-personal.99"},
		{"1.0.1-personal.1", "1.0.1-personal"},
	} {
		if compareVersions(tc.newer, tc.older) <= 0 || compareVersions(tc.older, tc.newer) >= 0 {
			t.Errorf("wrong update order: %s must be newer than %s", tc.newer, tc.older)
		}
	}
	for _, tc := range [][2]string{
		{"v1.0.1-personal.10", "1.0.1-personal.10"},
		{"1.0.1+build.1", "1.0.1+build.2"},
	} {
		if compareVersions(tc[0], tc[1]) != 0 {
			t.Errorf("equivalent versions caused an update: %v", tc)
		}
	}
}
