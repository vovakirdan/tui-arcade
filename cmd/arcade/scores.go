package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/vovakirdan/tui-arcade/internal/registry"
	"github.com/vovakirdan/tui-arcade/internal/storage"
)

var scoresCmd = &cobra.Command{
	Use:   "scores <game>",
	Short: "Show high scores for a game",
	Long: `Display the top 10 high scores for the specified game.

Examples:
  arcade scores flappy
  arcade scores dino`,
	Args: cobra.ExactArgs(1),
	Run:  runScores,
}

func runScores(cmd *cobra.Command, args []string) {
	gameID := args[0]

	// Check if game exists
	if !registry.Exists(gameID) {
		fmt.Fprintf(os.Stderr, "Error: unknown game %q\n", gameID)
		fmt.Fprintln(os.Stderr, "Run 'arcade list' to see available games.")
		os.Exit(1)
	}

	// Get game title
	game, err := registry.Create(gameID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating game: %v\n", err)
		os.Exit(1)
	}
	title := game.Title()

	// Open score storage
	store, err := storage.Open(flagDBPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening scores database: %v\n", err)
		os.Exit(1)
	}

	// Get top scores
	scores, err := store.TopScores(gameID, 10)
	if err != nil {
		store.Close()
		fmt.Fprintf(os.Stderr, "Error retrieving scores: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	// Display scores
	fmt.Printf("High Scores - %s\n", title)
	fmt.Println()

	if len(scores) == 0 {
		fmt.Println("No scores recorded yet.")
		fmt.Println()
		fmt.Printf("Play 'arcade play %s' to set the first high score!\n", gameID)
		return
	}

	// Print header
	fmt.Printf("  %-4s  %-10s  %s\n", "Rank", "Score", "Date")
	fmt.Printf("  %-4s  %-10s  %s\n", "----", "-----", "----")

	// Print scores
	for i, entry := range scores {
		dateStr := entry.CreatedAt.Format("2006-01-02 15:04")
		fmt.Printf("  %-4d  %-10d  %s\n", i+1, entry.Score, dateStr)
	}

	// Show high score
	fmt.Println()
	if len(scores) > 0 {
		highScore, err := store.HighScore(gameID)
		if err == nil {
			fmt.Printf("Best: %d\n", highScore)
		}
	}
}
