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
	newModel := newM.(*MainModel)

	if newModel.Focus != FocusPlaylist {
		t.Errorf("Expected focus to be Playlist after Tab, got %v", newModel.Focus)
	}

	// Test Tab again to switch back
	newM2, _ := newModel.Update(msg)
	newModel2 := newM2.(*MainModel)

	if newModel2.Focus != FocusLibrary {
		t.Errorf("Expected focus to be Library after second Tab, got %v", newModel2.Focus)
	}
}

func TestQuitConfirmation(t *testing.T) {
	m := NewModel(nil, nil)

	// Press 'q'
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
	m2, _ := m.Update(msg)
	m = m2.(*MainModel)

	if !m.ShowQuitConfirm {
		t.Errorf("Expected ShowQuitConfirm to be true after pressing 'q'")
	}

	// Press 'n' to cancel
	msgN := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")}
	m3, _ := m.Update(msgN)
	m = m3.(*MainModel)

	if m.ShowQuitConfirm {
		t.Errorf("Expected ShowQuitConfirm to be false after pressing 'n'")
	}

	// Press 'q' again
	m4, _ := m.Update(msg)
	m = m4.(*MainModel)

	// Press 'y' to quit
	msgY := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")}
	_, cmdY := m.Update(msgY)

	if cmdY == nil {
		t.Errorf("Expected a command after pressing 'y'")
	}
	// Note: verifying it's tea.Quit is hard without running it
}
