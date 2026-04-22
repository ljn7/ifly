// Package updater fetches release metadata from GitHub and verifies
// downloaded binaries against their sha256 sidecars.
package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type Asset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

type Release struct {
	Tag    string  `json:"tag_name"`
	Assets []Asset `json:"assets"`
}

func (r Release) AssetFor(name string) *Asset {
	for i := range r.Assets {
		if r.Assets[i].Name == name {
			return &r.Assets[i]
		}
	}
	return nil
}

type Client struct {
	baseURL string
	owner   string
	repo    string
	http    *http.Client
}

func NewClient(baseURL, owner, repo string, h *http.Client) *Client {
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}
	if h == nil {
		h = http.DefaultClient
	}
	return &Client{baseURL: baseURL, owner: owner, repo: repo, http: h}
}

func (c *Client) LatestRelease() (Release, error) {
	url := c.baseURL
	if !strings.HasPrefix(url, "http") {
		return Release{}, errors.New("invalid base url")
	}
	if strings.Contains(url, "api.github.com") {
		url = fmt.Sprintf("%s/repos/%s/%s/releases/latest", url, c.owner, c.repo)
	}
	resp, err := c.http.Get(url)
	if err != nil {
		return Release{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return Release{}, fmt.Errorf("github: %s", resp.Status)
	}
	var r Release
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return Release{}, err
	}
	return r, nil
}

func Download(url string, h *http.Client) ([]byte, error) {
	if h == nil {
		h = http.DefaultClient
	}
	resp, err := h.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("download %s: %s", url, resp.Status)
	}
	return io.ReadAll(resp.Body)
}

// VerifyChecksum parses a sha256 sidecar ("<hex>  <filename>\n") and compares.
func VerifyChecksum(data []byte, sidecar string) error {
	want := strings.TrimSpace(strings.SplitN(strings.TrimSpace(sidecar), " ", 2)[0])
	sum := sha256.Sum256(data)
	got := hex.EncodeToString(sum[:])
	if !strings.EqualFold(got, want) {
		return fmt.Errorf("sha256 mismatch: got %s want %s", got, want)
	}
	return nil
}

// IsNewer compares dotted-numeric versions with an optional leading "v".
// Non-numeric current (e.g. "dev") is treated as infinitely old.
func IsNewer(current, remote string) bool {
	cur := splitSemver(current)
	rem := splitSemver(remote)
	if cur == nil {
		return true
	}
	for i := 0; i < 3; i++ {
		var a, bv int
		if i < len(cur) {
			a = cur[i]
		}
		if i < len(rem) {
			bv = rem[i]
		}
		if a != bv {
			return bv > a
		}
	}
	return false
}

func splitSemver(v string) []int {
	v = strings.TrimPrefix(v, "v")
	parts := strings.Split(v, ".")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil
		}
		out = append(out, n)
	}
	return out
}
