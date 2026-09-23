package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSetupModelTransitions(t *testing.T) {
	m := NewSetupModel()
	if m.Step != StepRadioName {
		t.Fatalf("expected initial step to be StepRadioName, got %v", m.Step)
	}

	// 1. Enter station name
	m.TextInputBuf = "Test Radio"
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if m.Step != StepRadioID {
		t.Fatalf("expected step to be StepRadioID, got %v", m.Step)
	}
	if m.StationName != "Test Radio" {
		t.Errorf("expected station name 'Test Radio', got %s", m.StationName)
	}
	if m.StationID != "testradio" {
		t.Errorf("expected station ID 'testradio', got %s", m.StationID)
	}

	// 2. Confirm station ID
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.Step != StepMediaSource {
		t.Fatalf("expected step to be StepMediaSource, got %v", m.Step)
	}

	// 3. Select local music
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.Step != StepFolderSelect {
		t.Fatalf("expected step to be StepFolderSelect, got %v", m.Step)
	}
}
