package main

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var banner = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("212")).
	Render("planosh")

var rootCmd = &cobra.Command{
	Use:   "planosh",
	Short: banner + " — deterministic plan-to-code harness",
	Long:  banner + " transforms PRDs into deterministic plan.sh + harness pairs and calibrates each step.",
	SilenceUsage:  true,
	SilenceErrors: true,
}
