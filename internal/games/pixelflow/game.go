// Package pixelflow provides the PixelFlow conveyor puzzle game for the arcade.
package pixelflow

import (
	"math/rand"
	"path/filepath"
	"runtime"

	platformcore "github.com/vovakirdan/tui-arcade/internal/core"
	"github.com/vovakirdan/tui-arcade/internal/games/pixelflow/core"
	"github.com/vovakirdan/tui-arcade/internal/games/pixelflow/levels"
	"github.com/vovakirdan/tui-arcade/internal/registry"
)

// Game implements the PixelFlow puzzle game.
type Game struct {
	rng    *rand.Rand
	state  *core.State
	level  levels.Level
	loader *levels.Loader

	// Game mode
	levelIndex int
	allLevels  []levels.Level

	// Screen dimensions
	screenW int
	screenH int

	// Status
	tick         uint64
	score        int
	gameOver     bool
	won          bool
	paused       bool
	tooSmall     bool
	autoTickRate int // Ticks between auto simulation steps

	// Rendering offsets
	gridOffsetX int
	gridOffsetY int
	hudHeight   int
}

// Package-level variables for configuration
var (
	selectedStartLevel int
)

// SetStartLevel sets the starting level (1-indexed). 0 means start from beginning.
func SetStartLevel(level int) {
	selectedStartLevel = level
}

// GetStartLevel returns the currently selected start level.
func GetStartLevel() int {
	return selectedStartLevel
}

func init() {
	registry.Register("pixelflow", func() registry.Game {
		return New()
	})
}

// New creates a new PixelFlow game.
func New() *Game {
	return &Game{
		hudHeight:    3,
		autoTickRate: 6, // Auto-step every 6 ticks (~10 steps/sec at 60 FPS)
	}
}

// ID returns the game identifier.
func (g *Game) ID() string {
	return "pixelflow"
}

// Title returns the display name.
func (g *Game) Title() string {
	return "PixelFlow"
}

// getLevelsPath returns the path to levels directory.
func getLevelsPath() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}
	dir := filepath.Dir(filename)
	return filepath.Join(dir, "testdata", "levels")
}

// Reset initializes or restarts the game.
func (g *Game) Reset(cfg platformcore.RuntimeConfig) {
	g.rng = rand.New(rand.NewSource(cfg.Seed))
	g.screenW = cfg.ScreenW
	g.screenH = cfg.ScreenH
	g.tick = 0
	g.score = 0
	g.gameOver = false
	g.won = false
	g.paused = false

	// Load levels
	g.loader = levels.NewLoader(getLevelsPath())
	allLevels, err := g.loader.LoadAll()
	if err != nil || len(allLevels) == 0 {
		g.gameOver = true
		return
	}
	g.allLevels = allLevels

	// Apply selected start level
	if selectedStartLevel > 0 && selectedStartLevel <= len(allLevels) {
		g.levelIndex = selectedStartLevel - 1
		selectedStartLevel = 0 // Reset after use
	} else {
		g.levelIndex = 0
	}

	g.loadCurrentLevel()
}

// loadCurrentLevel loads the level at current levelIndex.
func (g *Game) loadCurrentLevel() {
	if g.levelIndex >= len(g.allLevels) {
		g.won = true
		g.gameOver = true
		return
	}

	g.level = g.allLevels[g.levelIndex]

	// Check minimum size
	minW := g.level.Width + 4  // Grid + rail + border
	minH := g.level.Height + 6 // Grid + rail + HUD + status
	if g.screenW < minW || g.screenH < minH {
		g.tooSmall = true
		return
	}
	g.tooSmall = false

	// Generate a deck for this level
	deck := g.generateDeck()

	// Create game state
	g.state = g.level.NewState(deck)

	// Calculate grid offset to center it
	displayW := g.level.Width + 2 // Grid + rail borders
	displayH := g.level.Height + 2
	g.gridOffsetX = (g.screenW - displayW) / 2
	g.gridOffsetY = g.hudHeight + (g.screenH-g.hudHeight-displayH)/2
}

// generateDeck creates a solvable deck for the current level.
func (g *Game) generateDeck() []core.Shooter {
	grid := g.level.ToGrid()
	rail := g.level.ToRail()

	// Use the deck generator with default params
	params := core.DefaultGenParams()
	params.Seed = uint64(g.rng.Int63())
	params.Capacity = g.level.Capacity
	params.MaxAmmoPerShooter = 4

	deck, err := core.GenerateDeckForGrid(grid, rail, params)
	if err != nil {
		// Fallback to simple deck
		deck = core.GenerateSimpleDeck(grid)
	}
	return deck
}

// Step advances the game by one tick.
func (g *Game) Step(input platformcore.InputFrame) platformcore.StepResult {
	g.tick++

	// Handle restart
	if input.Has(platformcore.ActionRestart) && g.gameOver {
		g.Reset(platformcore.RuntimeConfig{
			Seed:    g.rng.Int63(),
			ScreenW: g.screenW,
			ScreenH: g.screenH,
		})
		return platformcore.StepResult{State: g.State()}
	}

	// Handle pause toggle
	if input.Has(platformcore.ActionPause) {
		g.paused = !g.paused
	}

	// Don't process if game over, paused, or too small
	if g.gameOver || g.paused || g.tooSmall || g.state == nil {
		return platformcore.StepResult{State: g.State()}
	}

	// Handle manual launch with Space/Confirm
	if input.Has(platformcore.ActionConfirm) || input.Has(platformcore.ActionJump) {
		g.state.AutoLaunch()
	}

	// Auto-step simulation while there are active shooters
	if len(g.state.Active) > 0 && g.tick%uint64(g.autoTickRate) == 0 {
		result := g.state.StepTick()
		g.score += len(result.Removed) * 10

		// Auto-launch when possible
		g.state.AutoLaunch()
	}

	// Check win/lose conditions
	if g.state.IsWon() {
		g.levelIndex++
		if g.levelIndex >= len(g.allLevels) {
			g.won = true
			g.gameOver = true
		} else {
			g.loadCurrentLevel()
		}
	} else if g.state.IsLost() {
		g.gameOver = true
	}

	return platformcore.StepResult{State: g.State()}
}

// Render draws the game to the screen.
func (g *Game) Render(dst *platformcore.Screen) {
	dst.Clear()

	// Draw HUD
	g.renderHUD(dst)

	// Handle special states
	if g.tooSmall {
		g.renderOverlay(dst, "Window too small", "Resize to continue")
		return
	}

	if g.state == nil {
		g.renderOverlay(dst, "No levels found", "Check levels directory")
		return
	}

	// Draw game grid with rail
	g.renderGrid(dst)

	// Draw deck preview
	g.renderDeck(dst)

	// Draw overlays
	switch {
	case g.won:
		g.renderOverlay(dst, "You Win!", "All levels cleared!")
	case g.gameOver:
		g.renderOverlay(dst, "Game Over", "Press R to restart")
	case g.paused:
		g.renderOverlay(dst, "Paused", "Press P to continue")
	}
}

// renderHUD draws the top status bar.
func (g *Game) renderHUD(dst *platformcore.Screen) {
	var hud string
	if g.state != nil {
		hud = " PixelFlow | Score: " + itoa(g.score) +
			" | Level: " + itoa(g.levelIndex+1) + "/" + itoa(len(g.allLevels)) +
			" | Pixels: " + itoa(g.state.RemainingPixels()) +
			" | Active: " + itoa(len(g.state.Active)) + "/" + itoa(g.state.Capacity)
	} else {
		hud = " PixelFlow"
	}

	dst.DrawTextWithColor(0, 0, hud, platformcore.ColorCyan)

	// Separator
	for x := 0; x < dst.Width(); x++ {
		dst.SetWithColor(x, 1, '─', platformcore.ColorGray)
	}

	// Controls hint
	controls := " Space: Launch | P: Pause | R: Restart | Q: Quit"
	dst.DrawTextWithColor(0, 2, controls, platformcore.ColorGray)
}

// renderGrid draws the pixel grid with rail and shooters.
func (g *Game) renderGrid(dst *platformcore.Screen) {
	if g.state == nil {
		return
	}

	// Build map of active shooter positions
	activeByRail := make(map[int]core.ActiveShooter)
	for _, a := range g.state.Active {
		activeByRail[a.RailIndex] = a
	}

	displayW := g.state.Grid.W + 2
	displayH := g.state.Grid.H + 2

	for dy := 0; dy < displayH; dy++ {
		for dx := 0; dx < displayW; dx++ {
			screenX := g.gridOffsetX + dx
			screenY := g.gridOffsetY + dy

			if screenX < 0 || screenX >= dst.Width() || screenY < 0 || screenY >= dst.Height() {
				continue
			}

			// Grid coordinates
			gx := dx - 1
			gy := dy - 1

			// Check if this is a rail position (border)
			if dx == 0 || dx == g.state.Grid.W+1 || dy == 0 || dy == g.state.Grid.H+1 {
				railIdx := g.displayToRailIndex(dx, dy)
				if railIdx >= 0 {
					if shooter, ok := activeByRail[railIdx]; ok {
						// Active shooter on rail
						char, color := g.shooterRune(shooter)
						dst.SetWithColor(screenX, screenY, char, color)
					} else {
						dst.SetWithColor(screenX, screenY, '·', platformcore.ColorGray)
					}
				} else {
					dst.SetWithColor(screenX, screenY, '+', platformcore.ColorGray)
				}
			} else if gx >= 0 && gx < g.state.Grid.W && gy >= 0 && gy < g.state.Grid.H {
				// Inside grid
				cell := g.state.Grid.Get(core.C(gx, gy))
				if cell.Filled {
					char := cell.Color.Char()
					color := g.pixelflowColorToCore(cell.Color)
					dst.SetWithColor(screenX, screenY, char, color)
				} else {
					dst.SetWithColor(screenX, screenY, '.', platformcore.ColorGray)
				}
			}
		}
	}
}

// shooterRune returns the character and color for an active shooter.
func (g *Game) shooterRune(shooter core.ActiveShooter) (rune, platformcore.Color) {
	char := shooter.Color.LowerChar()
	if shooter.Dry {
		char = 'x' // Dry shooters show as 'x'
	}
	color := g.pixelflowColorToCore(shooter.Color)
	return char, color
}

// displayToRailIndex converts display coordinates to rail index.
func (g *Game) displayToRailIndex(dx, dy int) int {
	w := g.state.Grid.W
	h := g.state.Grid.H

	// Top edge (dy=0, dx=1 to w)
	if dy == 0 && dx >= 1 && dx <= w {
		return dx - 1
	}

	// Right edge (dx=w+1, dy=1 to h)
	if dx == w+1 && dy >= 1 && dy <= h {
		return w + (dy - 1)
	}

	// Bottom edge (dy=h+1, dx=w to 1) - reversed
	if dy == h+1 && dx >= 1 && dx <= w {
		return w + h + (w - dx)
	}

	// Left edge (dx=0, dy=h to 1) - reversed
	if dx == 0 && dy >= 1 && dy <= h {
		return 2*w + h + (h - dy)
	}

	return -1 // Corner
}

// pixelflowColorToCore maps pixelflow colors to platform core colors.
func (g *Game) pixelflowColorToCore(c core.Color) platformcore.Color {
	switch c {
	case core.ColorPink:
		return platformcore.ColorMagenta
	case core.ColorCyan:
		return platformcore.ColorCyan
	case core.ColorGreen:
		return platformcore.ColorGreen
	case core.ColorYellow:
		return platformcore.ColorYellow
	case core.ColorPurple:
		return platformcore.ColorBrightMagenta
	default:
		return platformcore.ColorWhite
	}
}

// renderDeck draws the deck preview below the grid.
func (g *Game) renderDeck(dst *platformcore.Screen) {
	if g.state == nil {
		return
	}

	deckY := g.gridOffsetY + g.state.Grid.H + 3
	if deckY >= dst.Height()-1 {
		return
	}

	dst.DrawTextWithColor(1, deckY, "Deck: ", platformcore.ColorGray)

	x := 7
	for i, shooter := range g.state.Deck {
		if i >= 8 {
			dst.DrawTextWithColor(x, deckY, "...", platformcore.ColorGray)
			break
		}
		char := shooter.Color.LowerChar()
		color := g.pixelflowColorToCore(shooter.Color)
		dst.SetWithColor(x, deckY, char, color)
		x++
		// Show ammo count
		ammoStr := itoa(shooter.Ammo)
		dst.DrawText(x, deckY, ammoStr)
		x += len(ammoStr) + 1
	}

	// Show waiting shooters
	if len(g.state.Waiting) > 0 {
		waitY := deckY + 1
		if waitY < dst.Height() {
			dst.DrawTextWithColor(1, waitY, "Wait: ", platformcore.ColorGray)
			x = 7
			for _, shooter := range g.state.Waiting {
				char := shooter.Color.LowerChar()
				color := g.pixelflowColorToCore(shooter.Color)
				dst.SetWithColor(x, waitY, char, color)
				x++
				ammoStr := itoa(shooter.Ammo)
				dst.DrawText(x, waitY, ammoStr)
				x += len(ammoStr) + 1
			}
		}
	}
}

// renderOverlay draws a centered overlay message.
func (g *Game) renderOverlay(dst *platformcore.Screen, line1, line2 string) {
	w := dst.Width()
	h := dst.Height()

	maxLen := len(line1)
	if len(line2) > maxLen {
		maxLen = len(line2)
	}
	boxW := maxLen + 4
	boxH := 5
	boxX := (w - boxW) / 2
	boxY := (h - boxH) / 2

	// Draw box
	for y := boxY; y < boxY+boxH && y < h; y++ {
		for x := boxX; x < boxX+boxW && x < w; x++ {
			if x < 0 || y < 0 {
				continue
			}
			isTopOrBottom := y == boxY || y == boxY+boxH-1
			isLeftOrRight := x == boxX || x == boxX+boxW-1
			switch {
			case isTopOrBottom && isLeftOrRight:
				dst.Set(x, y, '+')
			case isTopOrBottom:
				dst.Set(x, y, '-')
			case isLeftOrRight:
				dst.Set(x, y, '|')
			default:
				dst.Set(x, y, ' ')
			}
		}
	}

	// Draw text
	g.drawCenteredText(dst, line1, boxY+1)
	g.drawCenteredText(dst, line2, boxY+3)
}

func (g *Game) drawCenteredText(dst *platformcore.Screen, text string, y int) {
	if y < 0 || y >= dst.Height() {
		return
	}
	x := (dst.Width() - len(text)) / 2
	for i, ch := range text {
		px := x + i
		if px >= 0 && px < dst.Width() {
			dst.Set(px, y, ch)
		}
	}
}

// State returns the current game state.
func (g *Game) State() platformcore.GameState {
	return platformcore.GameState{
		Score:    g.score,
		GameOver: g.gameOver,
		Paused:   g.paused,
	}
}

// LevelCount returns the number of available levels.
func LevelCount() int {
	loader := levels.NewLoader(getLevelsPath())
	lvls, err := loader.LoadAll()
	if err != nil {
		return 0
	}
	return len(lvls)
}

// LevelNames returns the names of all levels.
func LevelNames() []string {
	loader := levels.NewLoader(getLevelsPath())
	lvls, err := loader.LoadAll()
	if err != nil {
		return nil
	}
	names := make([]string, len(lvls))
	for i, l := range lvls {
		names[i] = l.Name
	}
	return names
}

// itoa is a simple int to string converter.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	negative := n < 0
	if negative {
		n = -n
	}
	digits := make([]byte, 0, 10)
	for n > 0 {
		digits = append(digits, byte('0'+n%10))
		n /= 10
	}
	// Reverse
	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}
	if negative {
		return "-" + string(digits)
	}
	return string(digits)
}
