package main

import (
	"fmt"
	"os"

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

var (
	flagConfig     string
	flagDifficulty string
)

var playCmd = &cobra.Command{
	Use:   "play <game>",
	Short: "Play a game",
	Long: `Start playing the specified game.

Controls:
  Space/Up   - Jump/Flap
  P/Esc      - Pause
  R          - Restart (after game over)
  Q/Ctrl+C   - Quit

Difficulty options:
  easy   - Start at lowest difficulty, progresses to max
  normal - Start at 30% difficulty, progresses to max
  hard   - Start at 70% difficulty, progresses to max
  fixed  - No progression, stays at config's initial level

Examples:
  arcade play flappy
  arcade play dino --difficulty easy
  arcade play flappy --difficulty hard
  arcade play dino --difficulty fixed
  arcade play flappy --config ./my-flappy.yaml`,
	Args: cobra.ExactArgs(1),
	Run:  runPlay,
}

func init() {
	playCmd.Flags().StringVar(&flagConfig, "config", "", "Path to custom game config YAML")
	playCmd.Flags().StringVar(&flagDifficulty, "difficulty", "", "Difficulty preset: easy, normal, hard, fixed")
}

func runPlay(cmd *cobra.Command, args []string) {
	gameID := args[0]

	// Check if game exists
	if !registry.Exists(gameID) {
		fmt.Fprintf(os.Stderr, "Error: unknown game %q\n", gameID)
		fmt.Fprintln(os.Stderr, "Run 'arcade list' to see available games.")
		os.Exit(1)
	}

	// Get terminal size early for mode selector
	width, height := 80, 24 // Defaults
	if w, h, termErr := term.GetSize(int(os.Stdout.Fd())); termErr == nil {
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
		selection, updatedCfg, selErr := tui.RunBreakoutModeSelector(cfg)
		if selErr != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", selErr)
			os.Exit(1)
		}
		cfg = updatedCfg

		// User pressed back or quit
		if selection == nil {
			return
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
		snakeSelection, updatedCfg, snakeErr := tui.RunSnakeModeSelector(cfg)
		if snakeErr != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", snakeErr)
			os.Exit(1)
		}
		cfg = updatedCfg

		// User pressed back or quit
		if snakeSelection == nil {
			return
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
		t2048Selection, updatedCfg, t2048Err := tui.RunT2048ModeSelector(cfg)
		if t2048Err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", t2048Err)
			os.Exit(1)
		}
		cfg = updatedCfg

		// User pressed back or quit
		if t2048Selection == nil {
			return
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
		pfSelection, updatedCfg, pfErr := tui.RunPixelFlowLevelSelector(cfg)
		if pfErr != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", pfErr)
			os.Exit(1)
		}
		cfg = updatedCfg

		// User pressed back or quit
		if pfSelection == nil {
			return
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
		os.Exit(1)
	}

	// Open score storage
	store, err := storage.Open(flagDBPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not open scores database: %v\n", err)
		// Continue without storage - game still works
		store = nil
	}

	// Run the game
	runErr := tui.Run(game, store, cfg)

	// Close store before potential exit
	if store != nil {
		store.Close()
	}

	if runErr != nil {
		fmt.Fprintf(os.Stderr, "Error running game: %v\n", runErr)
		os.Exit(1)
	}
}
