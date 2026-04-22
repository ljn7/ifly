package egg

import (
	"strings"
	"testing"
)

func TestFooterNeverLovesIfEasterEggOff(t *testing.T) {
	for i := 0; i < 1000; i++ {
		f := Footer("0.1.0", false, i)
		if strings.Contains(f, "\U0001F49C") {
			t.Fatalf("easter egg off but got %q", f)
		}
	}
}

func TestFooterRotatesPool(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 200; i++ {
		seen[Footer("0.1.0", true, i)] = true
	}
	if len(seen) < 2 {
		t.Errorf("expected rotation, only saw %v", seen)
	}
}

func TestFooterLoveAtMultiplesOfTen(t *testing.T) {
	got := Footer("0.1.0", true, 0)
	if got != "\U0001F49C ifly" {
		t.Errorf("index 0 should be love footer, got %q", got)
	}
}
