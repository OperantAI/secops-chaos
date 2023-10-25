package output

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

func WriteInfo(msg string, args ...interface{}) {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	fmt.Println(style.Render(fmt.Sprintf(msg, args...)))
}

func WriteSuccess(msg string, args ...interface{}) {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	fmt.Println(style.Render(fmt.Sprintf(msg, args...)))
}

func WriteWarning(msg string, args ...interface{}) {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	fmt.Println(style.Render(fmt.Sprintf(msg, args...)))
}

func WriteError(err error) {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	fmt.Println(style.Render(err.Error()))
}

func WriteFatal(err error) {
	WriteError(err)
	os.Exit(1)
}

func WriteTable(headers []string, rows [][]string) {
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("205"))).
		Headers(headers...).
		Rows(rows...)

	fmt.Println(t)
}
