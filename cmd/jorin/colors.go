package main

import (
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Styles using lipgloss for consistent, composable styling.
	styleBold   = lipgloss.NewStyle().Bold(true)
	styleRed    = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff3b30"))
	styleGreen  = lipgloss.NewStyle().Foreground(lipgloss.Color("#34c759"))
	styleYellow = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffcc00"))
	styleBlue   = lipgloss.NewStyle().Foreground(lipgloss.Color("#007aff"))
	// Tool calls / debugging should be grey
	styleCyan = lipgloss.NewStyle().Foreground(lipgloss.Color("#9E9E9E"))

	// Additional styles for REPL usage
	// LLM responses: off-white / very light blue
	stylePrompt = lipgloss.NewStyle().Foreground(lipgloss.Color("#e6f7ff")).Bold(true)
	styleHeader = lipgloss.NewStyle().Foreground(lipgloss.Color("#e6f7ff")).Bold(true)
	// Info and tool output use grey for debugging
	styleInfo  = lipgloss.NewStyle().Foreground(lipgloss.Color("#9E9E9E"))
	styleError = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff3b30")).Bold(true)
)

var colorEnabled = detectColor()

func detectColor() bool {
	// Respect NO_COLOR
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if strings.ToLower(os.Getenv("TERM")) == "dumb" {
		return false
	}
	// Check if either stderr or stdout is a char device (terminal)
	if fi, err := os.Stderr.Stat(); err == nil {
		if fi.Mode()&os.ModeCharDevice != 0 {
			return true
		}
	}
	if fi, err := os.Stdout.Stat(); err == nil {
		if fi.Mode()&os.ModeCharDevice != 0 {
			return true
		}
	}
	return false
}

func colorize(style lipgloss.Style, s string) string {
	if !colorEnabled || s == "" {
		return s
	}
	return style.Render(s)
}

// Convenience wrappers for commonly used colorized strings.
func bold(s string) string           { return colorize(styleBold, s) }
func red(s string) string            { return colorize(styleRed, s) }
func green(s string) string          { return colorize(styleGreen, s) }
func yellow(s string) string         { return colorize(styleYellow, s) }
func blue(s string) string           { return colorize(styleBlue, s) }
func cyan(s string) string           { return colorize(styleCyan, s) }
func promptStyleStr(s string) string { return colorize(stylePrompt, s) }
func headerStyleStr(s string) string { return colorize(styleHeader, s) }
func infoStyleStr(s string) string   { return colorize(styleInfo, s) }
func errorStyleStr(s string) string  { return colorize(styleError, s) }

// Prevent "unused" linter errors for styles and helper functions that may
// currently be unused but are kept for readability and future use.
var _ = []interface{}{styleBold, styleRed, styleGreen, styleYellow, styleBlue, styleCyan, stylePrompt, styleHeader, styleInfo, styleError}
var _ = []interface{}{bold, red, green, yellow, blue, cyan, promptStyleStr, headerStyleStr, infoStyleStr, errorStyleStr}
