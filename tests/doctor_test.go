package tests

import (
	"testing"

	"aircoda/internal/cli"
)

func TestRunDoctor(t *testing.T) {
	err := cli.RunDoctor(false)
	if err != nil {
		t.Fatalf("RunDoctor failed: %v", err)
	}
}
