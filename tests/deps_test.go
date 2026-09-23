package tests

import (
	"testing"

	"aircoda/internal/deps"
)

func TestCheckDependencies(t *testing.T) {
	result, err := deps.CheckDependencies()
	if result == nil {
		t.Fatal("Expected non-nil result from CheckDependencies")
	}

	if !result.FFmpeg {
		if err == nil {
			t.Error("Expected error when FFmpeg is missing, got nil")
		}
	} else {
		t.Log("FFmpeg is installed and detected in system PATH")
	}
}
