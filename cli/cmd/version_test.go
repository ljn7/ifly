package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionPrintsBanner(t *testing.T) {
	SetVersion("0.1.0")
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"version"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "IFLy v0.1.0") {
		t.Errorf("missing banner: %q", out)
	}
	if !strings.Contains(out, "I'm Feeling Lucky") {
		t.Errorf("missing tagline: %q", out)
	}
}

func TestVersionLoveFlag(t *testing.T) {
	SetVersion("0.1.0")
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"version", "--love"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "I Fucking Love You") {
		t.Errorf("expected love banner, got %q", out)
	}
}
