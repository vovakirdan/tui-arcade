// arcade is a TUI arcade platform for playing retro-style games in the terminal.
//
// Usage:
//
//	arcade list              - List available games
//	arcade play <game>       - Play a game
//	arcade menu              - Start menu to pick games interactively
//	arcade serve             - Start SSH server for remote play
//	arcade scores <game>     - Show high scores for a game
//
// Global flags:
//
//	--fps <rate>    - Set tick rate (default: 60)
//	--seed <value>  - Set RNG seed for reproducible gameplay
//	--db <path>     - Set database path (default: ~/.arcade/scores.db)
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	// Import games to register them
	_ "github.com/vovakirdan/tui-arcade/internal/games/breakout"
	_ "github.com/vovakirdan/tui-arcade/internal/games/dino"
	_ "github.com/vovakirdan/tui-arcade/internal/games/flappy"
	_ "github.com/vovakirdan/tui-arcade/internal/games/pong"
	_ "github.com/vovakirdan/tui-arcade/internal/games/snake"
	_ "github.com/vovakirdan/tui-arcade/internal/games/t2048"
)

var (
	// Global flags
	flagFPS    int
	flagSeed   int64
	flagDBPath string
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "arcade",
	Short: "TUI Arcade - Play retro games in your terminal",
	Long: `TUI Arcade is a terminal-based gaming platform that lets you play
classic-style games directly in your terminal.

Available commands:
  list     - Show all available games
  play     - Play a specific game directly
  menu     - Interactive game picker menu
  serve    - Start SSH server for remote play
  scores   - View high scores

Examples:
  arcade list
  arcade play flappy
  arcade menu
  arcade serve --ssh :2222
  arcade scores flappy`,
}

func init() {
	// Global persistent flags
	rootCmd.PersistentFlags().IntVar(&flagFPS, "fps", 60, "Tick rate (frames per second)")
	rootCmd.PersistentFlags().Int64Var(&flagSeed, "seed", 0, "RNG seed (0 = random based on time)")
	rootCmd.PersistentFlags().StringVar(&flagDBPath, "db", "~/.arcade/scores.db", "Path to scores database")

	// Add subcommands
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(playCmd)
	rootCmd.AddCommand(menuCmd)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(scoresCmd)
}
