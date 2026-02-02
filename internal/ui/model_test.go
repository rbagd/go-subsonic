package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestUpdate(t *testing.T) {
	m := NewModel(nil, nil)

	// Initial focus should be Library
	if m.Focus != FocusLibrary {
		t.Errorf("Expected initial focus to be Library, got %v", m.Focus)
	}
	if !m.Library.IsFocused {
		t.Errorf("Expected Library to be focused")
	}

	// Test Tab to switch focus
	msg := tea.KeyMsg{Type: tea.KeyTab}
	newM, _ := m.Update(msg)
	newModel := newM.(MainModel)

	if newModel.Focus != FocusPlaylist {
		t.Errorf("Expected focus to be Playlist after Tab, got %v", newModel.Focus)
	}

	// Test Tab again to switch back
	newM2, _ := newModel.Update(msg)
	newModel2 := newM2.(MainModel)

	if newModel2.Focus != FocusLibrary {
		t.Errorf("Expected focus to be Library after second Tab, got %v", newModel2.Focus)
	}
}
