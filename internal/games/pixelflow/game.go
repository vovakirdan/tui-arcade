// Package pixelflow provides the PixelFlow conveyor puzzle game for the arcade.
package pixelflow

import (
	"math/rand"
	"path/filepath"
	"runtime"
	"strings"

	platformcore "github.com/vovakirdan/tui-arcade/internal/core"
	"github.com/vovakirdan/tui-arcade/internal/games/pixelflow/core"
	"github.com/vovakirdan/tui-arcade/internal/games/pixelflow/levels"
	"github.com/vovakirdan/tui-arcade/internal/registry"
)

// FocusArea indicates which area has input focus.
type FocusArea int

const (
	FocusDeck FocusArea = iota
	FocusWaiting
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

	// Selection state
	focus         FocusArea // Current focus (deck or waiting)
	selectedQueue int       // Selected deck queue index
	selectedSlot  int       // Selected waiting slot index

	// Rendering config
	cellW      int // Width of each grid cell in terminal chars
	cellH      int // Height of each grid cell in terminal lines
	railOffset int // Gap between grid and rail
	hudHeight  int

	// Calculated offsets
	gridOffsetX int
	gridOffsetY int
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
		hudHeight:    4,
		autoTickRate: 6,  // Auto-step every 6 ticks (~10 steps/sec at 60 FPS)
		cellW:        2,  // Each pixel is 2 chars wide
		cellH:        1,  // Each pixel is 1 line tall
		railOffset:   1,  // 1 cell gap between grid and rail
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
	g.focus = FocusDeck
	g.selectedQueue = 0
	g.selectedSlot = 0

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

	// Calculate cell dimensions based on available space
	g.calculateLayout()

	if g.tooSmall {
		return
	}

	// Generate a deck for this level
	deck := g.generateDeck()

	// Create game state
	g.state = g.level.NewState(deck)

	// Reset selection
	g.focus = FocusDeck
	g.selectedQueue = 0
	g.selectedSlot = 0
}

// calculateLayout determines cell sizes and offsets.
func (g *Game) calculateLayout() {
	// Available space after HUD and status areas
	availW := g.screenW - 4  // margins
	availH := g.screenH - g.hudHeight - 6 // HUD + deck display + margins

	// Grid dimensions with rail offset
	gridDisplayW := g.level.Width
	gridDisplayH := g.level.Height

	// Calculate maximum cell size that fits
	// Need space for: rail + offset + grid + offset + rail
	// Rail is 1 cell wide, offset is railOffset cells
	totalGridW := gridDisplayW + 2*(1+g.railOffset)
	totalGridH := gridDisplayH + 2*(1+g.railOffset)

	// Calculate cell dimensions
	g.cellW = availW / totalGridW
	g.cellH = availH / totalGridH

	// Enforce minimums
	if g.cellW < 2 {
		g.cellW = 2
	}
	if g.cellH < 1 {
		g.cellH = 1
	}

	// Check if we have enough space
	neededW := totalGridW * g.cellW
	neededH := totalGridH * g.cellH + g.hudHeight + 4

	if g.screenW < neededW || g.screenH < neededH {
		g.tooSmall = true
		return
	}
	g.tooSmall = false

	// Center the grid display
	displayW := totalGridW * g.cellW
	displayH := totalGridH * g.cellH
	g.gridOffsetX = (g.screenW - displayW) / 2
	g.gridOffsetY = g.hudHeight + (g.screenH - g.hudHeight - displayH - 4) / 2
}

// generateDeck creates a solvable deck for the current level.
func (g *Game) generateDeck() []core.Shooter {
	grid := g.level.ToGrid()
	rail := g.level.ToRail()

	// Use the deck generator with default params
	params := core.DefaultGenParams()
	params.Seed = uint64(g.rng.Int63())
	params.Capacity = g.level.Capacity
	params.NumQueues = g.level.NumQueues
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

	// Handle selection navigation
	if input.Has(platformcore.ActionUp) {
		g.focus = FocusDeck
	}
	if input.Has(platformcore.ActionDown) {
		g.focus = FocusWaiting
	}
	if input.Has(platformcore.ActionLeft) {
		if g.focus == FocusDeck {
			if g.selectedQueue > 0 {
				g.selectedQueue--
			}
		} else {
			if g.selectedSlot > 0 {
				g.selectedSlot--
			}
		}
	}
	if input.Has(platformcore.ActionRight) {
		if g.focus == FocusDeck {
			if g.selectedQueue < g.state.NumQueues-1 {
				g.selectedQueue++
			}
		} else {
			if g.selectedSlot < g.state.Waiting.Capacity-1 {
				g.selectedSlot++
			}
		}
	}

	// Handle manual launch with Space/Confirm
	if input.Has(platformcore.ActionConfirm) || input.Has(platformcore.ActionJump) {
		if g.focus == FocusDeck {
			g.state.LaunchFromQueue(g.selectedQueue)
		} else {
			g.state.LaunchFromWaiting(g.selectedSlot)
		}
	}

	// Auto-step simulation while there are active shooters
	if len(g.state.Active) > 0 && g.tick%uint64(g.autoTickRate) == 0 {
		result := g.state.StepTick()
		g.score += len(result.Removed) * 10
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

	// Draw game grid with rail (scaled colored blocks)
	g.renderScaledGrid(dst)

	// Draw deck queues and waiting slots
	g.renderQueuesAndSlots(dst)

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
	var controls string
	if g.focus == FocusDeck {
		controls = " [DECK] ←/→: Queue | ↑/↓: Focus | Space: Launch | P: Pause"
	} else {
		controls = " [WAIT] ←/→: Slot | ↑/↓: Focus | Space: Launch | P: Pause"
	}
	dst.DrawTextWithColor(0, 2, controls, platformcore.ColorGray)

	// Another separator
	for x := 0; x < dst.Width(); x++ {
		dst.SetWithColor(x, 3, '─', platformcore.ColorGray)
	}
}

// renderScaledGrid draws the pixel grid with colored blocks and rail.
func (g *Game) renderScaledGrid(dst *platformcore.Screen) {
	if g.state == nil {
		return
	}

	// Build map of active shooter positions
	activeByRail := make(map[int]core.ActiveShooter)
	for _, a := range g.state.Active {
		activeByRail[a.RailIndex] = a
	}

	// Total display dimensions including rail and offset
	railAndOffset := 1 + g.railOffset
	totalW := g.state.Grid.W + 2*railAndOffset
	totalH := g.state.Grid.H + 2*railAndOffset

	// Draw each cell
	for dy := 0; dy < totalH; dy++ {
		for dx := 0; dx < totalW; dx++ {
			// Screen position for this cell
			screenX := g.gridOffsetX + dx*g.cellW
			screenY := g.gridOffsetY + dy*g.cellH

			// Grid coordinates (accounting for rail and offset)
			gx := dx - railAndOffset
			gy := dy - railAndOffset

			// Determine what to draw
			isRail := (dx == 0 || dx == totalW-1 || dy == 0 || dy == totalH-1)
			isOffset := !isRail && (dx < railAndOffset || dx >= totalW-railAndOffset ||
				dy < railAndOffset || dy >= totalH-railAndOffset)
			isGrid := !isRail && !isOffset

			if isRail {
				// Rail position
				railIdx := g.displayToRailIndex(dx, dy, totalW, totalH, railAndOffset)
				if railIdx >= 0 {
					if shooter, ok := activeByRail[railIdx]; ok {
						// Active shooter on rail
						g.renderRailShooter(dst, screenX, screenY, shooter)
					} else {
						// Empty rail
						g.renderRailEmpty(dst, screenX, screenY)
					}
				} else {
					// Corner
					g.renderRailCorner(dst, screenX, screenY)
				}
			} else if isOffset {
				// Offset area (empty space between rail and grid)
				g.renderEmpty(dst, screenX, screenY)
			} else if isGrid && gx >= 0 && gx < g.state.Grid.W && gy >= 0 && gy < g.state.Grid.H {
				// Grid cell
				cell := g.state.Grid.Get(core.C(gx, gy))
				if cell.Filled {
					g.renderPixelBlock(dst, screenX, screenY, cell.Color)
				} else {
					g.renderEmptyCell(dst, screenX, screenY)
				}
			}
		}
	}
}

// renderPixelBlock draws a colored block for a pixel.
func (g *Game) renderPixelBlock(dst *platformcore.Screen, x, y int, c core.Color) {
	color := g.pixelflowColorToCore(c)
	// Draw block as colored spaces (background color)
	blockChar := '█' // Full block character
	for cy := 0; cy < g.cellH; cy++ {
		for cx := 0; cx < g.cellW; cx++ {
			if x+cx >= 0 && x+cx < dst.Width() && y+cy >= 0 && y+cy < dst.Height() {
				dst.SetWithColor(x+cx, y+cy, blockChar, color)
			}
		}
	}
}

// renderEmptyCell draws an empty grid cell.
func (g *Game) renderEmptyCell(dst *platformcore.Screen, x, y int) {
	for cy := 0; cy < g.cellH; cy++ {
		for cx := 0; cx < g.cellW; cx++ {
			if x+cx >= 0 && x+cx < dst.Width() && y+cy >= 0 && y+cy < dst.Height() {
				dst.SetWithColor(x+cx, y+cy, '·', platformcore.ColorGray)
			}
		}
	}
}

// renderEmpty draws empty space.
func (g *Game) renderEmpty(dst *platformcore.Screen, x, y int) {
	for cy := 0; cy < g.cellH; cy++ {
		for cx := 0; cx < g.cellW; cx++ {
			if x+cx >= 0 && x+cx < dst.Width() && y+cy >= 0 && y+cy < dst.Height() {
				dst.Set(x+cx, y+cy, ' ')
			}
		}
	}
}

// renderRailEmpty draws an empty rail position.
func (g *Game) renderRailEmpty(dst *platformcore.Screen, x, y int) {
	for cy := 0; cy < g.cellH; cy++ {
		for cx := 0; cx < g.cellW; cx++ {
			if x+cx >= 0 && x+cx < dst.Width() && y+cy >= 0 && y+cy < dst.Height() {
				dst.SetWithColor(x+cx, y+cy, '░', platformcore.ColorGray)
			}
		}
	}
}

// renderRailCorner draws a rail corner.
func (g *Game) renderRailCorner(dst *platformcore.Screen, x, y int) {
	for cy := 0; cy < g.cellH; cy++ {
		for cx := 0; cx < g.cellW; cx++ {
			if x+cx >= 0 && x+cx < dst.Width() && y+cy >= 0 && y+cy < dst.Height() {
				dst.SetWithColor(x+cx, y+cy, '╬', platformcore.ColorGray)
			}
		}
	}
}

// renderRailShooter draws a shooter on the rail.
func (g *Game) renderRailShooter(dst *platformcore.Screen, x, y int, shooter core.ActiveShooter) {
	color := g.pixelflowColorToCore(shooter.Color)
	char := '●'
	if shooter.Dry {
		char = '○'
	}
	for cy := 0; cy < g.cellH; cy++ {
		for cx := 0; cx < g.cellW; cx++ {
			if x+cx >= 0 && x+cx < dst.Width() && y+cy >= 0 && y+cy < dst.Height() {
				dst.SetWithColor(x+cx, y+cy, char, color)
			}
		}
	}
}

// displayToRailIndex converts display coordinates to rail index.
func (g *Game) displayToRailIndex(dx, dy, totalW, totalH, railAndOffset int) int {
	w := g.state.Grid.W
	h := g.state.Grid.H

	// Only outer border is rail
	if dy == 0 {
		// Top edge
		gridX := dx - railAndOffset
		if gridX >= 0 && gridX < w {
			return gridX
		}
	} else if dx == totalW-1 {
		// Right edge
		gridY := dy - railAndOffset
		if gridY >= 0 && gridY < h {
			return w + gridY
		}
	} else if dy == totalH-1 {
		// Bottom edge (reversed)
		gridX := dx - railAndOffset
		if gridX >= 0 && gridX < w {
			return w + h + (w - 1 - gridX)
		}
	} else if dx == 0 {
		// Left edge (reversed)
		gridY := dy - railAndOffset
		if gridY >= 0 && gridY < h {
			return 2*w + h + (h - 1 - gridY)
		}
	}

	return -1 // Corner or invalid
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

// renderQueuesAndSlots draws the deck queues and waiting slots.
func (g *Game) renderQueuesAndSlots(dst *platformcore.Screen) {
	if g.state == nil {
		return
	}

	// Calculate Y position for queues area
	railAndOffset := 1 + g.railOffset
	totalH := g.state.Grid.H + 2*railAndOffset
	baseY := g.gridOffsetY + totalH*g.cellH + 1

	// Draw deck queues
	dst.DrawTextWithColor(1, baseY, "Queues:", platformcore.ColorGray)
	queueY := baseY + 1

	for qi := 0; qi < g.state.Deck.NumQueues(); qi++ {
		x := 2 + qi*12
		if x >= dst.Width()-10 {
			break
		}

		// Queue header with selection indicator
		header := "Q" + itoa(qi+1) + ":"
		if g.focus == FocusDeck && qi == g.selectedQueue {
			header = "[" + header + "]"
			dst.DrawTextWithColor(x-1, queueY, header, platformcore.ColorBrightYellow)
		} else {
			dst.DrawTextWithColor(x, queueY, header, platformcore.ColorGray)
		}

		// Show top 3 shooters in queue
		for si := 0; si < 3 && si < g.state.Deck.QueueLen(qi); si++ {
			shooter := g.state.Deck.Queues[qi][si]
			char := shooter.Color.LowerChar()
			color := g.pixelflowColorToCore(shooter.Color)
			dst.SetWithColor(x+4+si*3, queueY, char, color)
			dst.DrawText(x+5+si*3, queueY, itoa(shooter.Ammo))
		}
		if g.state.Deck.QueueLen(qi) > 3 {
			dst.DrawTextWithColor(x+13, queueY, "...", platformcore.ColorGray)
		}
	}

	// Draw waiting slots
	waitY := queueY + 2
	dst.DrawTextWithColor(1, waitY, "Waiting:", platformcore.ColorGray)
	slotY := waitY + 1

	for si := 0; si < g.state.Waiting.Capacity; si++ {
		x := 2 + si*6
		if x >= dst.Width()-5 {
			break
		}

		slot := g.state.Waiting.Get(si)
		// Slot indicator
		if g.focus == FocusWaiting && si == g.selectedSlot {
			dst.DrawTextWithColor(x-1, slotY, "[", platformcore.ColorBrightYellow)
			dst.DrawTextWithColor(x+4, slotY, "]", platformcore.ColorBrightYellow)
		}

		if slot != nil {
			char := slot.Color.LowerChar()
			color := g.pixelflowColorToCore(slot.Color)
			dst.SetWithColor(x, slotY, char, color)
			dst.DrawText(x+1, slotY, itoa(slot.Ammo))
		} else {
			dst.DrawTextWithColor(x, slotY, "___", platformcore.ColorGray)
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

// Ensure strings package is used
var _ = strings.Repeat
