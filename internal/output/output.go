package output

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

const (
	InfoColor        = lipgloss.Color("205")
	SuccessColor     = lipgloss.Color("205")
	WarningColor     = lipgloss.Color("205")
	ErrorColor       = lipgloss.Color("205")
	FatalColor       = lipgloss.Color("205")
	TableBorderColor = lipgloss.Color("205")
)

func WriteInfo(msg string, args ...interface{}) {
	style := lipgloss.NewStyle().Foreground(InfoColor)
	fmt.Println(style.Render(fmt.Sprintf(msg, args...)))
}

func WriteSuccess(msg string, args ...interface{}) {
	style := lipgloss.NewStyle().Foreground(SuccessColor)
	fmt.Println(style.Render(fmt.Sprintf(msg, args...)))
}

func WriteWarning(msg string, args ...interface{}) {
	style := lipgloss.NewStyle().Foreground(WarningColor)
	fmt.Println(style.Render(fmt.Sprintf(msg, args...)))
}

func WriteError(err error) {
	style := lipgloss.NewStyle().Foreground(ErrorColor)
	fmt.Println(style.Render(err.Error()))
}

func WriteFatal(err error) {
	WriteError(err)
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
