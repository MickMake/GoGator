package gps

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"gogator/internal/config"
)

func TestReadRawCSVRawRowsMatchSourceLines(t *testing.T) {
	loc := time.UTC
	cfg := config.Default()
	cfg.RawTime.CorrectionHours = 0

	dir := t.TempDir()
	headed := filepath.Join(dir, "headed.csv")
	if err := os.WriteFile(headed, []byte("dt,lat,lng,altitude,angle,speed,params\n2026-04-01 00:00:00,-33.1,151.1,0,0,0,io24=0\n"), 0644); err != nil {
		t.Fatal(err)
	}
	points, err := ReadRawCSV(headed, loc, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(points) != 1 || points[0].RawRow != 2 {
		t.Fatalf("headed raw row = %#v, want first data row RawRow 2", points)
	}

	headerless := filepath.Join(dir, "headerless.csv")
	if err := os.WriteFile(headerless, []byte("2026-04-01 00:00:00,-33.1,151.1,0,0,0,io24=0\n"), 0644); err != nil {
		t.Fatal(err)
	}
	points, err = ReadRawCSV(headerless, loc, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(points) != 1 || points[0].RawRow != 1 {
		t.Fatalf("headerless raw row = %#v, want first data row RawRow 1", points)
	}
}
