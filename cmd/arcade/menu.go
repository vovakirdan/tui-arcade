package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/vovakirdan/tui-arcade/internal/core"
	"github.com/vovakirdan/tui-arcade/internal/games/breakout"
	"github.com/vovakirdan/tui-arcade/internal/games/dino"
	"github.com/vovakirdan/tui-arcade/internal/games/flappy"
	"github.com/vovakirdan/tui-arcade/internal/games/pixelflow"
	"github.com/vovakirdan/tui-arcade/internal/games/snake"
	"github.com/vovakirdan/tui-arcade/internal/games/t2048"
	"github.com/vovakirdan/tui-arcade/internal/platform/tui"
	"github.com/vovakirdan/tui-arcade/internal/registry"
	"github.com/vovakirdan/tui-arcade/internal/storage"
)

var menuCmd = &cobra.Command{
	Use:   "menu",
	Short: "Start the arcade with a game picker menu",
	Long: `Start the arcade in interactive menu mode.

Use arrow keys or j/k to navigate, Enter to select a game.
After a game ends, you return to the menu to play again.

Controls:
  Up/Down/j/k  - Navigate menu
  Enter/Space  - Select game
  Q            - Quit

Examples:
  arcade menu
  arcade menu --fps 30
  arcade menu --db ./scores.db`,
	Run: runMenu,
}

func init() {
	// Uses global flags from main.go (--fps, --seed, --db)
}

func runMenu(_ *cobra.Command, _ []string) {
	// Open score storage
	store, err := storage.Open(flagDBPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not open scores database: %v\n", err)
		store = nil
	}

	// Get terminal size
	width, height := 80, 24
	if w, h, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		width = w
		height = h
	}

	// Create runtime config
	cfg := core.RuntimeConfig{
		ScreenW:  width,
		ScreenH:  height,
		TickRate: flagFPS,
		Seed:     flagSeed,
	}

	// Menu loop
	for {
		// Show menu and get selection
		menuResult, err := tui.RunMenu(store, cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			break
		}

		// Update config with any size changes
		cfg = menuResult.Config

		// Check if user quit
		if menuResult.Quit {
			break
		}

		// Check if user wants scoreboard
		if menuResult.WantsScoreboard {
			goBack, sbErr := tui.RunScoreboard(store, cfg.ScreenW, cfg.ScreenH)
			if sbErr != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", sbErr)
			}
			if goBack {
				continue // Back to menu
			}
			break // User quit from scoreboard
		}

		gameID := menuResult.GameID
		if gameID == "" {
			break
		}

		// Set config path and difficulty for games before creation
		switch gameID {
		case "flappy":
			flappy.SetConfigPath(flagConfig)
			flappy.SetDifficultyPreset(flagDifficulty)
		case "dino":
			dino.SetConfigPath(flagConfig)
			dino.SetDifficultyPreset(flagDifficulty)
		case "breakout":
			breakout.SetConfigPath(flagConfig)
			breakout.SetDifficultyPreset(flagDifficulty)

			// Show Breakout mode/level selector
			selection, updatedCfg2, breakoutErr := tui.RunBreakoutModeSelector(cfg)
			if breakoutErr != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", breakoutErr)
				continue
			}
			cfg = updatedCfg2

			// User pressed back or quit
			if selection == nil {
				continue
			}

			// Apply selection
			if selection.Mode == tui.BreakoutModeEndless {
				gameID = "breakout_endless"
			}
			if selection.Level > 0 {
				breakout.SetStartLevel(selection.Level)
			}

		case "snake":
			snake.SetConfigPath(flagConfig)
			snake.SetDifficultyPreset(flagDifficulty)

			// Show Snake mode/level selector
			snakeSelection, updatedCfg3, snakeErr := tui.RunSnakeModeSelector(cfg)
			if snakeErr != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", snakeErr)
				continue
			}
			cfg = updatedCfg3

			// User pressed back or quit
			if snakeSelection == nil {
				continue
			}

			// Apply selection
			if snakeSelection.Mode == tui.SnakeModeEndless {
				gameID = "snake_endless"
			}
			if snakeSelection.Level > 0 {
				snake.SetStartLevel(snakeSelection.Level)
			}

		case "2048":
			// Show 2048 mode/level selector
			t2048Selection, updatedCfg4, t2048Err := tui.RunT2048ModeSelector(cfg)
			if t2048Err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", t2048Err)
				continue
			}
			cfg = updatedCfg4

			// User pressed back or quit
			if t2048Selection == nil {
				continue
			}

			// Apply selection
			if t2048Selection.Mode == tui.T2048ModeEndless {
				gameID = "2048_endless"
			}
			if t2048Selection.Level > 0 {
				t2048.SetStartLevel(t2048Selection.Level)
			}

		case "pixelflow":
			// Show PixelFlow level selector
			pfSelection, updatedCfg5, pfErr := tui.RunPixelFlowLevelSelector(cfg)
			if pfErr != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", pfErr)
				continue
			}
			cfg = updatedCfg5

			// User pressed back or quit
			if pfSelection == nil {
				continue
			}

			// Apply selection
			if pfSelection.Level > 0 {
				pixelflow.SetStartLevel(pfSelection.Level)
			}
		}

		// Create game instance
		game, err := registry.Create(gameID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating game: %v\n", err)
			continue
		}

		// Update seed for each game
		cfg.Seed = time.Now().UnixNano()

		// Run the game
		if err := tui.Run(game, store, cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error running game: %v\n", err)
		}

		// Loop back to menu
	}

	// Cleanup
	if store != nil {
		store.Close()
	}
}
