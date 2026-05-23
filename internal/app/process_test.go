package app

import (
	"context"
	"errors"
	"strings"
	"testing"

	"gogator/engine"
	"gogator/internal/config"
)

func TestRunProcessPipelinePropagatesEngineError(t *testing.T) {
	t.Parallel()
	orig := runEngine
	t.Cleanup(func() { runEngine = orig })

	runEngine = func(_ context.Context, _ engine.Input) (engine.Result, error) {
		return engine.Result{}, errors.New("boom")
	}

	_, err := runProcessPipeline(nil, nil, nil, config.Default())
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "run engine pipeline") {
		t.Fatalf("expected wrapped engine error, got %v", err)
	}
}
