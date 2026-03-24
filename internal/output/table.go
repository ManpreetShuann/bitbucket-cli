package output

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

type TableFormatter struct {
	columns []Column
	noColor bool
}

func NewTableFormatter(columns []Column, noColor bool) *TableFormatter {
	return &TableFormatter{columns: columns, noColor: noColor}
}

func (f *TableFormatter) FormatRows(rows [][]string) string {
	if len(rows) == 0 {
		return "No results found."
	}

	isTTY := term.IsTerminal(int(os.Stdout.Fd()))
	useColor := isTTY && !f.noColor

	var sb strings.Builder

	headerStyle := lipgloss.NewStyle().Bold(true)
	for i, col := range f.columns {
		header := col.Header
		if useColor {
			header = headerStyle.Render(header)
		}
		if i < len(f.columns)-1 {
			sb.WriteString(fmt.Sprintf("%-*s  ", col.Width, header))
		} else {
			sb.WriteString(header)
		}
	}
	sb.WriteString("\n")

	for _, row := range rows {
		for i, cell := range row {
			if i < len(f.columns)-1 {
				truncated := truncate(cell, f.columns[i].Width)
				sb.WriteString(fmt.Sprintf("%-*s  ", f.columns[i].Width, truncated))
			} else {
				sb.WriteString(cell)
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// IsTerminal reports whether stdout is a terminal.
func IsTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}
