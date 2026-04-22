package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLatestReleaseParsesTag(t *testing.T) {
	body := `{"tag_name":"v0.2.0","assets":[{"name":"ifly-linux-amd64","browser_download_url":"http://x/y"},{"name":"ifly-linux-amd64.sha256","browser_download_url":"http://x/y.sha256"}]}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, body)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "ljn7", "ifly", &http.Client{})
	rel, err := c.LatestRelease()
	if err != nil {
		t.Fatal(err)
	}
	if rel.Tag != "v0.2.0" {
		t.Errorf("tag %q", rel.Tag)
	}
	if rel.AssetFor("ifly-linux-amd64") == nil {
		t.Error("expected asset for ifly-linux-amd64")
	}
}

func TestIsNewerSemver(t *testing.T) {
	cases := []struct {
		current, remote string
		want            bool
	}{
		{"0.1.0", "0.2.0", true},
		{"v0.1.0", "v0.1.1", true},
		{"0.2.0", "0.2.0", false},
		{"0.3.0", "0.2.9", false},
		{"dev", "0.1.0", true},
	}
	for _, c := range cases {
		if got := IsNewer(c.current, c.remote); got != c.want {
			t.Errorf("IsNewer(%q,%q)=%v want %v", c.current, c.remote, got, c.want)
		}
	}
}

func TestVerifyChecksum(t *testing.T) {
	data := []byte("hello")
	sum := sha256.Sum256(data)
	h := hex.EncodeToString(sum[:])
	if err := VerifyChecksum(data, h+"  ifly-linux-amd64\n"); err != nil {
		t.Errorf("unexpected err: %v", err)
	}
	if err := VerifyChecksum(data, "deadbeef"); err == nil {
		t.Error("expected checksum mismatch")
	}
}

func TestDownloadReturnsBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "binary-bytes")
	}))
	defer srv.Close()
	data, err := Download(srv.URL, &http.Client{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(data), "binary") {
		t.Errorf("got %q", data)
	}
}
