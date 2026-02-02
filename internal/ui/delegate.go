package ui

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// item is a wrapper for list items
type item struct {
	title    string
	desc     string
	id       string
	isDir    bool
	duration int
	// Metadata
	artist string
	album  string
	year   int
	track  int
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

// CustomDelegate handles rendering of list items
type CustomDelegate struct{}

func (d CustomDelegate) Height() int  { return 1 }
func (d CustomDelegate) Spacing() int { return 0 }
func (d CustomDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d CustomDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := i.title

	// If it's an artist (has description "X Albums"), format it specially
	if strings.Contains(i.desc, "Albums") {
		count := i.desc
		if m.Index() == index {
			fmt.Fprintf(w, "%s %s", selectedItemStyle.Render(str), selectedItemStyle.Copy().Faint(true).Render(count))
		} else {
			fmt.Fprintf(w, "%s %s", itemStyle.Render(str), faintStyle.Render(count))
		}
	} else {
		// Normal item (Album/Song)
		var content string
		if !i.isDir && i.duration > 0 {
			dur := FmtDuration(time.Duration(i.duration) * time.Second)
			content = fmt.Sprintf("%s %s", str, faintStyle.Render(dur))
		} else {
			content = str
		}

		if m.Index() == index {
			fmt.Fprintf(w, "%s", selectedItemStyle.Render(content))
		} else {
			fmt.Fprintf(w, "%s", itemStyle.Render(content))
		}
	}
}

func FmtDuration(d time.Duration) string {
	d = d.Round(time.Second)
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	return fmt.Sprintf("%02d:%02d", m, s)
}
