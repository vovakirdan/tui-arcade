package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/vovakirdan/tui-arcade/internal/registry"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available games",
	Long:  `Shows a list of all games registered in the arcade.`,
	Run:   runList,
}

func runList(cmd *cobra.Command, args []string) {
	games := registry.List()

	if len(games) == 0 {
		fmt.Println("No games available.")
		return
	}

	fmt.Println("Available games:")
	fmt.Println()

	// Calculate column widths
	maxIDLen := 2 // "ID" header
	for _, g := range games {
		if len(g.ID) > maxIDLen {
			maxIDLen = len(g.ID)
		}
	}

	// Print header
	fmt.Printf("  %-*s  %s\n", maxIDLen, "ID", "Title")
	fmt.Printf("  %-*s  %s\n", maxIDLen, "--", "-----")

	// Print games
	for _, g := range games {
		fmt.Printf("  %-*s  %s\n", maxIDLen, g.ID, g.Title)
	}

	fmt.Println()
	fmt.Println("Run 'arcade play <id>' to play a game.")
}
