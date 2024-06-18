package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	ltable "github.com/charmbracelet/lipgloss/table"
	"github.com/mattn/go-runewidth"
	"golang.org/x/term"
	"gopkg.in/yaml.v2"
)

const (
	InfoColor    = lipgloss.Color("86")
	SuccessColor = lipgloss.Color("78")
	WarningColor = lipgloss.Color("208")
	ErrorColor   = lipgloss.Color("196")
	FatalColor   = lipgloss.Color("196")
)

func WriteInfo(msg string, args ...interface{}) {
	style := lipgloss.NewStyle().Foreground(InfoColor)
	fmt.Printf("%s %s\n", style.Render("INFO"), fmt.Sprintf(msg, args...))
}

func WriteSuccess(msg string, args ...interface{}) {
	style := lipgloss.NewStyle().Foreground(SuccessColor)
	fmt.Printf("%s %s\n", style.Render("SUCCESS"), fmt.Sprintf(msg, args...))
}

func WriteWarning(msg string, args ...interface{}) {
	style := lipgloss.NewStyle().Foreground(WarningColor)
	fmt.Printf("%s %s\n", style.Render("WARN"), fmt.Sprintf(msg, args...))
}

func WriteError(msg string, args ...interface{}) {
	style := lipgloss.NewStyle().Foreground(ErrorColor)
	fmt.Printf("%s %s\n", style.Render("ERROR"), fmt.Sprintf(msg, args...))
}

func WriteFatal(msg string, args ...interface{}) {
	style := lipgloss.NewStyle().Foreground(ErrorColor)
	fmt.Printf("%s %s\n", style.Render("FATAL"), fmt.Sprintf(msg, args...))
	os.Exit(1)
}

type table struct {
	headers []string
	rows    [][]string
}

// NewTable creates a new table with the given headers
func NewTable(headers []string) *table {
	return &table{headers: headers}
}

// AddRow adds a row to the table
func (t *table) AddRow(row []string) {
	t.rows = append(t.rows, row)
}

// Render renders the table, printing it to stdout
func (t *table) Render() {
	physicalWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		fmt.Println("Error getting terminal size:", err)
		return
	}

	// Calculate max width for each column
	columnCount := len(t.headers)
	columnWidth := physicalWidth / columnCount

	// Truncate cells if they exceed the column width
	truncatedRows := make([][]string, len(t.rows))
	for i, row := range t.rows {
		truncatedRows[i] = make([]string, columnCount)
		for j, cell := range row {
			truncatedRows[i][j] = truncateString(cell, columnWidth)
		}
	}

	// Create the table using lipgloss
	tbl := ltable.New().
		StyleFunc(func(row, col int) lipgloss.Style {
			if col == 1 {
				return lipgloss.NewStyle().MaxWidth(physicalWidth)
			}
			return lipgloss.Style{}
		}).
		Border(lipgloss.NormalBorder()).
		Headers(t.headers...).
		Rows(truncatedRows...)

	fmt.Println(tbl)
}

// truncateString ensures the string fits within the given width, appending "..." if necessary
func truncateString(str string, width int) string {
	if runewidth.StringWidth(str) <= width {
		return str
	}
	truncated := runewidth.Truncate(str, width-3, "...")
	return truncated
}

// WriteJSON writes the given output as pretty printed JSON to stdout
func WriteJSON(output interface{}) {
	jsonOutput, err := json.MarshalIndent(output, "", "    ")
	if err != nil {
		WriteError("Failed to marshal JSON: %s", err)
	}
	fmt.Println(string(jsonOutput))
}

// WriteYAML writes the given output as a pretty printed YAML object to stdout
func WriteYAML(output interface{}) {
	yamlOutput, err := yaml.Marshal(output)
	if err != nil {
		WriteError("Failed to marshal YAML: %s", err)
	}
	fmt.Println(string(yamlOutput))
}
