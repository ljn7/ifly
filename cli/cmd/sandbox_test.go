package cmd

import (
	"bytes"
	"runtime"
	"strings"
	"testing"
)

func TestSandboxCommandExists(t *testing.T) {
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"sandbox", "--help"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "namespace") {
		t.Errorf("expected help text to mention namespace: %q", buf.String())
	}
}

func TestSandboxNonLinuxPrintsGuidance(t *testing.T) {
	if runtime.GOOS == "linux" {
		t.Skip()
	}
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"sandbox", "--", "claude"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error on non-linux")
	}
	if !strings.Contains(err.Error(), "Linux") && !strings.Contains(err.Error(), "linux") {
		t.Errorf("error should mention Linux: %v", err)
	}
}
