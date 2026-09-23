package picker

import (
	"testing"
)

type mockPicker struct {
	folder string
	err    error
}

func (m *mockPicker) PickFolder() (string, error) {
	return m.folder, m.err
}

func TestMockFolderPicker(t *testing.T) {
	mock := &mockPicker{folder: "/tmp/music", err: nil}
	res, err := mock.PickFolder()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res != "/tmp/music" {
		t.Errorf("expected /tmp/music, got %s", res)
	}
}

func TestSystemFolderPickerInit(t *testing.T) {
	p := NewSystemFolderPicker()
	if p == nil {
		t.Fatal("expected non-nil SystemFolderPicker")
	}
}
