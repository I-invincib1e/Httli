package styles

import (
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
)

// Styles holds all the color styles for terminal output
type Styles struct {
	Header  lipgloss.Style
	Method  lipgloss.Style
	Status  lipgloss.Style
	Success lipgloss.Style
	Error   lipgloss.Style
	URL     lipgloss.Style
	Key     lipgloss.Style
	Value   lipgloss.Style
	Body    lipgloss.Style
}

// colorStyle creates a lipgloss style with the given color
func colorStyle(color string, bold bool, padding ...int) lipgloss.Style {
	s := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
	if bold {
		s = s.Bold(true)
	}
	if len(padding) >= 2 {
		s = s.Padding(padding[0], padding[1])
	}
	return s
}

// IsTTY returns true if stdout is a terminal (not piped/redirected)
func IsTTY() bool {
	return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
}

// New creates and returns a new Styles instance with all styles initialized.
// When stdout is not a TTY (piped or redirected), returns plain styles with no ANSI codes.
func New() *Styles {
	if !IsTTY() {
		// Return zero-value styles — lipgloss renders as plain text
		return &Styles{
			Header:  lipgloss.NewStyle(),
			Method:  lipgloss.NewStyle(),
			Status:  lipgloss.NewStyle(),
			Success: lipgloss.NewStyle(),
			Error:   lipgloss.NewStyle(),
			URL:     lipgloss.NewStyle(),
			Key:     lipgloss.NewStyle(),
			Value:   lipgloss.NewStyle(),
			Body:    lipgloss.NewStyle(),
		}
	}

	return &Styles{
		Header:  colorStyle("205", true, 0, 1),
		Method:  colorStyle("86", true),
		Status:  colorStyle("39", true),
		Success: colorStyle("46", false),
		Error:   colorStyle("196", false),
		URL:     colorStyle("33", false),
		Key:     colorStyle("214", false),
		Value:   colorStyle("252", false),
		Body:    colorStyle("252", false, 0, 1),
	}
}

// StatusColor returns the appropriate color code for a given HTTP status code.
// Returns empty string when not a TTY so no ANSI codes are emitted.
func StatusColor(statusCode int) string {
	if !IsTTY() {
		return ""
	}
	if statusCode >= 200 && statusCode < 300 {
		return "46" // green
	} else if statusCode >= 300 && statusCode < 400 {
		return "226" // yellow
	}
	return "196" // red
}
