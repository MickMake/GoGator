package store

import (
	"path/filepath"
	"testing"
)

func TestInitIsIdempotentAndStatusCounts(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "gogator.sqlite")

	if err := Init(path); err != nil {
		t.Fatalf("init first: %v", err)
	}
	if err := Init(path); err != nil {
		t.Fatalf("init second: %v", err)
	}

	counts, version, err := Status(path)
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if version == "" {
		t.Fatalf("expected sqlite version")
	}
	if counts.GPSPoints != 0 || counts.Sites != 0 || counts.Routes != 0 || counts.ProcessingRuns != 0 || counts.Trips != 0 || counts.Issues != 0 {
		t.Fatalf("unexpected counts: %+v", counts)
	}
}
