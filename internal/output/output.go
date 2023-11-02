package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	ltable "github.com/charmbracelet/lipgloss/table"
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
	fmt.Printf("%s %s", style.Render("INFO"), fmt.Sprintf(msg, args...))
}

func WriteSuccess(msg string, args ...interface{}) {
	style := lipgloss.NewStyle().Foreground(SuccessColor)
	fmt.Printf("%s %s", style.Render("SUCCESS"), fmt.Sprintf(msg, args...))
}

func WriteWarning(msg string, args ...interface{}) {
	style := lipgloss.NewStyle().Foreground(WarningColor)
	fmt.Printf("%s %s", style.Render("WARN"), fmt.Sprintf(msg, args...))
}

func WriteError(msg string, args ...interface{}) {
	style := lipgloss.NewStyle().Foreground(ErrorColor)
	fmt.Printf("%s %s", style.Render("ERROR"), fmt.Sprintf(msg, args...))
}

func WriteFatal(msg string, args ...interface{}) {
	style := lipgloss.NewStyle().Foreground(ErrorColor)
	fmt.Printf("%s %s", style.Render("FATAL"), fmt.Sprintf(msg, args...))
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
	tbl := ltable.New().
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case col == 1:
				return lipgloss.NewStyle().Width(40)
			}
			return lipgloss.Style{}
		}).
		Border(lipgloss.NormalBorder()).
		Headers(t.headers...).
		Rows(t.rows...)

	fmt.Println(tbl)
}

// WriteJSON writes the given output as pretty printed JSON to stdout
func WriteJSON(output interface{}) {
	jsonOutput, err := json.MarshalIndent(output, "", "    ")
	if err != nil {
		WriteError("Failed to marshal JSON: %s", err)
	}
	fmt.Println(string(jsonOutput))
}
