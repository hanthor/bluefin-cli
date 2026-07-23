package update

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsNewer(t *testing.T) {
	cases := []struct {
		current, latest string
		want            bool
	}{
		{"0.6.4", "v0.7.0", true},
		{"0.7.0", "v0.7.0", false},
		{"0.8.0", "v0.7.0", false},
		{"dev", "v99.0.0", false},
		{"v0.6.4", "0.7.0", true},
		{"garbage", "v0.7.0", false},
	}
	for _, c := range cases {
		if got := IsNewer(c.current, c.latest); got != c.want {
			t.Errorf("IsNewer(%q, %q) = %v, want %v", c.current, c.latest, got, c.want)
		}
	}
}

func checksumRelease(t *testing.T, body string) *Release {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, body)
	}))
	t.Cleanup(srv.Close)
	rel := &Release{TagName: "v1.0.0"}
	rel.Assets = append(rel.Assets, struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	}{Name: "checksums.txt", BrowserDownloadURL: srv.URL})
	return rel
}

func TestVerifyChecksum(t *testing.T) {
	data := []byte("archive-bytes")
	sum := sha256.Sum256(data)
	good := hex.EncodeToString(sum[:])

	rel := checksumRelease(t, good+"  app_1.0.0_linux_amd64.tar.gz\n")
	if err := verifyChecksum(rel, "app_1.0.0_linux_amd64.tar.gz", data); err != nil {
		t.Errorf("valid checksum rejected: %v", err)
	}

	rel = checksumRelease(t, good+"  app_1.0.0_linux_amd64.tar.gz\n")
	if err := verifyChecksum(rel, "app_1.0.0_linux_amd64.tar.gz", []byte("tampered")); err == nil {
		t.Error("tampered archive accepted")
	}

	rel = checksumRelease(t, good+"  some_other_asset.tar.gz\n")
	if err := verifyChecksum(rel, "app_1.0.0_linux_amd64.tar.gz", data); err == nil {
		t.Error("missing checksum entry accepted")
	}

	if err := verifyChecksum(&Release{TagName: "v1.0.0"}, "x.tar.gz", data); err == nil {
		t.Error("release without checksums.txt accepted")
	}
}
