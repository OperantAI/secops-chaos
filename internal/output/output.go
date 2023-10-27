package output

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

const (
	InfoColor        = lipgloss.Color("86")
	SuccessColor     = lipgloss.Color("78")
	WarningColor     = lipgloss.Color("208")
	ErrorColor       = lipgloss.Color("196")
	FatalColor       = lipgloss.Color("196")
	TableBorderColor = lipgloss.Color("205")
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

func WriteTable(headers []string, rows [][]string) {
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(TableBorderColor)).
		Headers(headers...).
		Rows(rows...)

	fmt.Println(t)
}
