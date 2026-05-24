package output

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gogator/engine"
)

func TestWriteEngineDiagnosticsMotionSpeed(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "x_engine_motion.csv")
	err := WriteEngineDiagnostics(engine.Diagnostics{Motion: []engine.MotionSample{{Index: 7, SpeedKPH: 42.5}}}, EngineDiagnosticPaths{Motion: p}, EngineDiagnosticOptions{Enabled: true, OutputMotion: true})
	if err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(p)
	text := string(b)
	if !strings.Contains(text, "speed_kmh") || !strings.Contains(text, "42.50") {
		t.Fatalf("unexpected motion diagnostics: %s", text)
	}
}
