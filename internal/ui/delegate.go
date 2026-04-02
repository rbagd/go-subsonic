package ui

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
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
type CustomDelegate struct {
	Focused   bool
	PlayingID string
}

func (d CustomDelegate) Height() int                               { return 1 }
func (d CustomDelegate) Spacing() int                              { return 0 }
func (d CustomDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d CustomDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := i.title
	isSelected := m.Index() == index
	activeStyle := itemStyle
	selectedStyle := selectedItemStyle
	if !d.Focused {
		selectedStyle = unfocusedSelectedItemStyle
	}

	// Special rendering for artists
	if m.Title == "Artists" {
		content := str
		itemRenderStyle := activeStyle
		selectedRenderStyle := selectedStyle
		if i.id == recentlyAddedNodeID {
			itemRenderStyle = recentItemStyle
			selectedRenderStyle = selectedRecentItemStyle
		} else {
			itemRenderStyle = activeStyle.Copy().Bold(true)
			selectedRenderStyle = selectedStyle.Copy().Bold(true)
		}

		if isSelected {
			fmt.Fprintf(w, "%s", selectedRenderStyle.Render(content))
		} else {
			fmt.Fprintf(w, "%s", itemRenderStyle.Render(content))
		}
	} else {
		// Normal item (Album/Song)
		content := str
		itemRenderStyle := activeStyle
		selectedRenderStyle := selectedStyle

		if m.Title == "Albums" || m.Title == "Recent Albums" {
			itemRenderStyle = activeStyle.Copy().Bold(true)
			selectedRenderStyle = selectedStyle.Copy().Bold(true)
		}

		if m.Title == "Songs" && !i.isDir && i.track != 0 {
			content = fmt.Sprintf("%d. %s", i.track, str)
		}

		switch {
		case m.Title == "Playlist" && !i.isDir:
			dur := ""
			if i.duration > 0 {
				dur = FmtDuration(time.Duration(i.duration) * time.Second)
			}

			marker := ""
			if d.PlayingID != "" && i.id == d.PlayingID {
				marker = "> "
			}

			year := ""
			if i.year != 0 {
				year = fmt.Sprintf("%d", i.year)
			}

			title := marker + content
			if d.PlayingID != "" && i.id == d.PlayingID {
				title = boldStyle.Render(title)
			}

			content = renderPlaylistLine(m.Width(), title, i.artist, i.album, year, dur)

		case !i.isDir && i.duration > 0:
			dur := FmtDuration(time.Duration(i.duration) * time.Second)
			content = fmt.Sprintf("%s %s", content, faintStyle.Render(dur))
		}

		if isSelected {
			fmt.Fprintf(w, "%s", selectedRenderStyle.Render(content))
		} else {
			fmt.Fprintf(w, "%s", itemRenderStyle.Render(content))
		}
	}
}

func renderPlaylistLine(width int, title, artist, album, year, dur string) string {
	const gap = 2

	if width <= 0 {
		return title
	}

	padding := itemStyle.GetPaddingLeft() + itemStyle.GetPaddingRight()
	available := width - padding
	if available <= 0 {
		return title
	}

	durWidth := ansi.StringWidth(dur)
	if dur == "" {
		durWidth = 0
	}

	yearWidth := ansi.StringWidth(year)
	if year == "" {
		yearWidth = 0
	}

	remain := available - durWidth - gap - (yearWidth + gap)
	if remain < 10 {
		// Fallback to two columns if too narrow
		meta := ""
		if artist != "" || album != "" {
			meta = fmt.Sprintf("(%s • %s)", artist, album)
		}
		if year != "" {
			meta = fmt.Sprintf("%s %s", meta, year)
		}
		return renderTwoColumnLine(width, title, meta, dur)
	}

	// Columns: Title (40%), Artist (20%), Album (20%)
	titleW := int(float64(remain) * 0.4)
	artistW := int(float64(remain) * 0.2)
	albumW := remain - titleW - artistW - gap*2

	if albumW < 5 {
		// Adjust if too narrow
		titleW = int(float64(remain) * 0.6)
		artistW = remain - titleW - gap
		albumW = 0
	}

	titlePart := ansi.Truncate(title, titleW, "…")
	artistPart := ""
	if artistW > 0 {
		artistPart = faintStyle.Render(ansi.Truncate(artist, artistW, "…"))
	}
	albumPart := ""
	if albumW > 0 {
		albumPart = faintStyle.Render(ansi.Truncate(album, albumW, "…"))
	}
	yearPart := ""
	if yearWidth > 0 {
		yearPart = faintStyle.Render(year)
	}
	durPart := ""
	if durWidth > 0 {
		durPart = faintStyle.Render(dur)
	}

	// Join with spaces to ensure alignment
	res := titlePart + strings.Repeat(" ", titleW+gap-ansi.StringWidth(titlePart))
	if artistW > 0 {
		res += artistPart + strings.Repeat(" ", artistW+gap-ansi.StringWidth(artistPart))
	}
	if albumW > 0 {
		res += albumPart + strings.Repeat(" ", albumW+gap-ansi.StringWidth(albumPart))
	}

	// Add year column
	if yearWidth > 0 {
		res += yearPart + strings.Repeat(" ", yearWidth+gap-ansi.StringWidth(yearPart))
	}

	// Right align duration
	spaces := available - ansi.StringWidth(res) - durWidth
	if spaces < 0 {
		spaces = 0
	}
	res += strings.Repeat(" ", spaces) + durPart

	return res
}

func renderTwoColumnLine(width int, title, meta, dur string) string {
	const gap = 1

	if width <= 0 {
		return title
	}

	padding := itemStyle.GetPaddingLeft() + itemStyle.GetPaddingRight()
	available := width - padding
	if available <= 0 {
		return title
	}

	leftPlain := strings.TrimSpace(title)
	if meta != "" {
		leftPlain = strings.TrimSpace(leftPlain + " " + meta)
	}

	rightPlain := strings.TrimSpace(dur)

	rightWidth := ansi.StringWidth(rightPlain)
	if rightPlain != "" && rightWidth >= available {
		return ansi.Truncate(rightPlain, available, "…")
	}

	leftBudget := available
	if rightPlain != "" {
		leftBudget = available - rightWidth - gap
		if leftBudget < 1 {
			leftBudget = 1
		}
	}

	leftTrunc := ansi.Truncate(leftPlain, leftBudget, "…")

	leftStyled := leftTrunc
	if meta != "" {
		metaIdx := strings.LastIndex(leftTrunc, meta)
		if metaIdx >= 0 {
			leftStyled = leftTrunc[:metaIdx] + faintStyle.Render(leftTrunc[metaIdx:])
		}
	}

	if rightPlain == "" {
		return leftStyled
	}

	leftWidth := ansi.StringWidth(leftTrunc)
	spaces := available - leftWidth - rightWidth
	if spaces < gap {
		spaces = gap
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, leftStyled, strings.Repeat(" ", spaces), faintStyle.Render(rightPlain))
}

func FmtDuration(d time.Duration) string {
	d = d.Round(time.Second)
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	return fmt.Sprintf("%02d:%02d", m, s)
}
