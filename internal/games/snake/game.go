package snake

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/vovakirdan/tui-arcade/internal/core"
	"github.com/vovakirdan/tui-arcade/internal/registry"
)

// Direction represents the snake's movement direction.
type Direction int

const (
	DirRight Direction = iota
	DirDown
	DirLeft
	DirUp
)

// Mode represents the game mode.
type Mode string

const (
	ModeCampaign Mode = "campaign"
	ModeEndless  Mode = "endless"
)

// Point represents a 2D coordinate.
type Point struct {
	X, Y int
}

// Game implements the Snake game.
type Game struct {
	mode           Mode
	rng            *rand.Rand
	tick           uint64
	score          int
	foodEaten      int // Food eaten in current level
	levelIndex     int // Current level (0-indexed)
	moveEveryTicks int
	moveTicker     int // Counts ticks until next move

	// Snake state
	snake     []Point // Head at index 0
	direction Direction
	nextDir   Direction // Buffered direction for next move
	growing   bool      // If true, don't remove tail on next move

	// Map state
	mapWidth   int
	mapHeight  int
	walls      map[Point]bool
	food       Point
	hudHeight  int
	mapOffsetX int
	mapOffsetY int

	// Screen dimensions
	screenW int
	screenH int

	// Game state flags
	gameOver     bool
	levelCleared bool
	won          bool
	paused       bool
	tooSmall     bool

	// Level clear animation
	levelClearTicks int
}

// Package-level variables for config/difficulty (like breakout pattern)
var (
	configPath         string //nolint:unused // kept for API compatibility
	difficultyPreset   string //nolint:unused // kept for API compatibility
	selectedStartLevel int
)

// SetConfigPath sets the config file path (for compatibility).
func SetConfigPath(path string) {
	configPath = path
}

// SetDifficultyPreset sets the difficulty preset (for compatibility).
func SetDifficultyPreset(preset string) {
	difficultyPreset = preset
}

// SetStartLevel sets the starting level (1-10). 0 means start from beginning.
func SetStartLevel(level int) {
	selectedStartLevel = level
}

// GetStartLevel returns the currently selected start level.
func GetStartLevel() int {
	return selectedStartLevel
}

// New creates a new campaign mode Snake game.
func New() *Game {
	return &Game{
		mode: ModeCampaign,
	}
}

// NewEndless creates a new endless mode Snake game.
func NewEndless() *Game {
	return &Game{
		mode: ModeEndless,
	}
}

func init() {
	registry.Register("snake", func() registry.Game {
		return New()
	})
	registry.Register("snake_endless", func() registry.Game {
		return NewEndless()
	})
}

// ID returns the game identifier.
func (g *Game) ID() string {
	if g.mode == ModeEndless {
		return "snake_endless"
	}
	return "snake"
}

// Title returns the display name.
func (g *Game) Title() string {
	if g.mode == ModeEndless {
		return "Snake (Endless)"
	}
	return "Snake"
}

// Reset initializes/restarts the game.
func (g *Game) Reset(cfg core.RuntimeConfig) {
	g.rng = rand.New(rand.NewSource(cfg.Seed))
	g.tick = 0
	g.score = 0
	g.foodEaten = 0
	g.gameOver = false
	g.levelCleared = false
	g.won = false
	g.paused = false
	g.tooSmall = false
	g.levelClearTicks = 0
	g.screenW = cfg.ScreenW
	g.screenH = cfg.ScreenH
	g.hudHeight = 2 // Top HUD lines

	// Apply selected start level (campaign only)
	if g.mode == ModeCampaign && selectedStartLevel > 0 && selectedStartLevel <= LevelCount() {
		g.levelIndex = selectedStartLevel - 1
		selectedStartLevel = 0 // Reset after use
	} else {
		g.levelIndex = 0
	}

	g.loadLevel()
}

// loadLevel loads the current level's map and spawns the snake.
func (g *Game) loadLevel() {
	level := GetLevel(g.levelIndex % LevelCount())
	if level == nil {
		return
	}

	// In endless mode, increase speed each cycle
	g.moveEveryTicks = level.MoveEveryTicks
	if g.mode == ModeEndless {
		cycle := g.levelIndex / LevelCount()
		// Each cycle reduces move interval by 1, minimum 1
		speedBonus := cycle
		g.moveEveryTicks = max(1, level.MoveEveryTicks-speedBonus)
	}
	g.moveTicker = 0
	g.foodEaten = 0
	g.levelCleared = false

	// Parse layout
	g.walls = make(map[Point]bool)
	layout := level.Layout
	g.mapHeight = len(layout)
	g.mapWidth = 0
	for _, row := range layout {
		if len(row) > g.mapWidth {
			g.mapWidth = len(row)
		}
	}

	// Check if screen is too small
	requiredW := g.mapWidth + 2
	requiredH := g.mapHeight + g.hudHeight + 1
	if g.screenW < requiredW || g.screenH < requiredH {
		g.tooSmall = true
		return
	}
	g.tooSmall = false

	// Center the map
	g.mapOffsetX = (g.screenW - g.mapWidth) / 2
	g.mapOffsetY = g.hudHeight

	// Parse walls
	for y, row := range layout {
		for x, ch := range row {
			if ch == '#' {
				g.walls[Point{X: x, Y: y}] = true
			}
		}
	}

	// Initialize snake in a safe starting position
	g.initSnake()

	// Spawn initial food
	g.spawnFood()
}

// initSnake places the snake at a safe starting position.
func (g *Game) initSnake() {
	// Find a good starting position (center-ish, not on walls)
	startX := g.mapWidth / 4
	startY := g.mapHeight / 2

	// Search for a clear spot
	for range 100 {
		clear := true
		for i := range 3 {
			p := Point{X: startX + i, Y: startY}
			if g.walls[p] || p.X < 1 || p.X >= g.mapWidth-1 || p.Y < 1 || p.Y >= g.mapHeight-1 {
				clear = false
				break
			}
		}
		if clear {
			break
		}
		// Try another position
		startX = 2 + g.rng.Intn(g.mapWidth/2)
		startY = 2 + g.rng.Intn(g.mapHeight-4)
	}

	// Create initial snake (3 segments, head at front)
	g.snake = []Point{
		{X: startX + 2, Y: startY}, // Head
		{X: startX + 1, Y: startY},
		{X: startX, Y: startY},
	}
	g.direction = DirRight
	g.nextDir = DirRight
	g.growing = false
}

// spawnFood places food at a random empty cell.
func (g *Game) spawnFood() {
	// Collect all empty cells
	var emptyCells []Point
	for y := 1; y < g.mapHeight-1; y++ {
		for x := 1; x < g.mapWidth-1; x++ {
			p := Point{X: x, Y: y}
			if !g.walls[p] && !g.isSnakeAt(p) {
				emptyCells = append(emptyCells, p)
			}
		}
	}

	if len(emptyCells) == 0 {
		// No space for food - should not happen in normal gameplay
		g.food = Point{X: -1, Y: -1}
		return
	}

	g.food = emptyCells[g.rng.Intn(len(emptyCells))]
}

// isSnakeAt checks if the snake occupies the given point.
func (g *Game) isSnakeAt(p Point) bool {
	for _, seg := range g.snake {
		if seg == p {
			return true
		}
	}
	return false
}

// Step advances the game by one tick.
func (g *Game) Step(input core.InputFrame) core.StepResult {
	g.tick++

	// Handle restart
	if input.Has(core.ActionRestart) && (g.gameOver || g.won) {
		g.Reset(core.RuntimeConfig{
			Seed:    g.rng.Int63(),
			ScreenW: g.screenW,
			ScreenH: g.screenH,
		})
		return core.StepResult{State: g.State()}
	}

	// Handle pause toggle
	if input.Has(core.ActionPause) {
		g.paused = !g.paused
	}

	// Don't process if game over, paused, too small, or level cleared animation
	if g.gameOver || g.won || g.paused || g.tooSmall {
		return core.StepResult{State: g.State()}
	}

	// Handle level cleared animation
	if g.levelCleared {
		g.levelClearTicks++
		if g.levelClearTicks >= 90 { // ~1.5 seconds at 60 FPS
			g.advanceLevel()
		}
		return core.StepResult{State: g.State()}
	}

	// Process direction input (buffer for next move)
	g.processInput(input)

	// Move snake on tick interval
	g.moveTicker++
	if g.moveTicker >= g.moveEveryTicks {
		g.moveTicker = 0
		g.moveSnake()
	}

	return core.StepResult{State: g.State()}
}

// processInput handles direction changes.
func (g *Game) processInput(input core.InputFrame) {
	newDir := g.nextDir

	switch {
	case input.Has(core.ActionUp):
		newDir = DirUp
	case input.Has(core.ActionDown):
		newDir = DirDown
	case input.Has(core.ActionLeft):
		newDir = DirLeft
	case input.Has(core.ActionRight):
		newDir = DirRight
	}

	// Prevent instant reversal
	if !g.isOpposite(newDir, g.direction) {
		g.nextDir = newDir
	}
}

// isOpposite checks if two directions are opposite.
func (g *Game) isOpposite(d1, d2 Direction) bool {
	return (d1 == DirUp && d2 == DirDown) ||
		(d1 == DirDown && d2 == DirUp) ||
		(d1 == DirLeft && d2 == DirRight) ||
		(d1 == DirRight && d2 == DirLeft)
}

// moveSnake moves the snake one cell in the current direction.
func (g *Game) moveSnake() {
	if len(g.snake) == 0 {
		return
	}

	// Apply buffered direction
	g.direction = g.nextDir

	// Calculate new head position
	head := g.snake[0]
	var newHead Point
	switch g.direction {
	case DirUp:
		newHead = Point{X: head.X, Y: head.Y - 1}
	case DirDown:
		newHead = Point{X: head.X, Y: head.Y + 1}
	case DirLeft:
		newHead = Point{X: head.X - 1, Y: head.Y}
	case DirRight:
		newHead = Point{X: head.X + 1, Y: head.Y}
	}

	// Check wall collision
	if g.walls[newHead] || newHead.X < 0 || newHead.X >= g.mapWidth ||
		newHead.Y < 0 || newHead.Y >= g.mapHeight {
		g.gameOver = true
		return
	}

	// Check self collision (excluding tail if not growing, since it will move)
	checkLen := len(g.snake)
	if !g.growing && checkLen > 0 {
		checkLen-- // Tail will be removed
	}
	for i := range checkLen {
		if g.snake[i] == newHead {
			g.gameOver = true
			return
		}
	}

	// Move snake: add new head
	g.snake = append([]Point{newHead}, g.snake...)

	// Check food collision
	if newHead == g.food {
		g.score++
		g.foodEaten++
		g.growing = true // Don't remove tail this move
		g.spawnFood()
		g.checkLevelCompletion()
	}

	// Remove tail unless growing
	if g.growing {
		g.growing = false
	} else if len(g.snake) > 1 {
		g.snake = g.snake[:len(g.snake)-1]
	}
}

// checkLevelCompletion checks if the level is complete.
func (g *Game) checkLevelCompletion() {
	if g.mode == ModeCampaign {
		level := GetLevel(g.levelIndex)
		if level != nil && g.foodEaten >= level.TargetFood {
			g.levelCleared = true
			g.levelClearTicks = 0
		}
	}
	// Endless mode: transition levels after every 10 food
	if g.mode == ModeEndless && g.foodEaten >= 10 {
		g.levelIndex++
		g.loadLevel()
	}
}

// advanceLevel moves to the next level.
func (g *Game) advanceLevel() {
	g.levelIndex++
	if g.mode == ModeCampaign && g.levelIndex >= LevelCount() {
		g.won = true
	} else {
		g.loadLevel()
	}
}

// Render draws the game to the screen.
func (g *Game) Render(dst *core.Screen) {
	dst.Clear()

	// Draw HUD
	g.renderHUD(dst)

	// Handle special states
	if g.tooSmall {
		g.renderOverlay(dst, "Window too small", "Resize to continue")
		return
	}

	// Draw map
	g.renderMap(dst)

	// Draw snake
	g.renderSnake(dst)

	// Draw food
	if g.food.X >= 0 && g.food.Y >= 0 {
		fx := g.mapOffsetX + g.food.X
		fy := g.mapOffsetY + g.food.Y
		if fx >= 0 && fx < dst.Width() && fy >= 0 && fy < dst.Height() {
			dst.Set(fx, fy, '*')
		}
	}

	// Draw overlays
	switch {
	case g.levelCleared:
		levelName := "Level"
		if level := GetLevel(g.levelIndex); level != nil {
			levelName = level.Name
		}
		g.renderOverlay(dst, fmt.Sprintf("Level %d cleared!", g.levelIndex+1), levelName)
	case g.won:
		g.renderOverlay(dst, "You Win!", fmt.Sprintf("Final Score: %d", g.score))
	case g.gameOver:
		g.renderOverlay(dst, "Game Over", "Press R to restart")
	case g.paused:
		g.renderOverlay(dst, "Paused", "Press P to continue")
	}
}

// renderHUD draws the top status bar.
func (g *Game) renderHUD(dst *core.Screen) {
	var hud string
	if g.mode == ModeEndless {
		hud = fmt.Sprintf(" Snake (Endless) — Score: %d  Speed: %d", g.score, 7-g.moveEveryTicks)
	} else {
		hud = fmt.Sprintf(" Snake — Score: %d  Level: %d/%d  Food: %d", g.score, g.levelIndex+1, LevelCount(), g.foodEaten)
	}

	// Draw HUD line
	for x := 0; x < dst.Width() && x < len(hud); x++ {
		dst.Set(x, 0, rune(hud[x]))
	}

	// Draw separator
	for x := range dst.Width() {
		dst.Set(x, 1, '─')
	}
}

// renderMap draws walls.
func (g *Game) renderMap(dst *core.Screen) {
	for wall := range g.walls {
		wx := g.mapOffsetX + wall.X
		wy := g.mapOffsetY + wall.Y
		if wx >= 0 && wx < dst.Width() && wy >= 0 && wy < dst.Height() {
			dst.Set(wx, wy, '#')
		}
	}
}

// renderSnake draws the snake.
func (g *Game) renderSnake(dst *core.Screen) {
	for i, seg := range g.snake {
		sx := g.mapOffsetX + seg.X
		sy := g.mapOffsetY + seg.Y
		if sx >= 0 && sx < dst.Width() && sy >= 0 && sy < dst.Height() {
			if i == 0 {
				dst.Set(sx, sy, 'O') // Head
			} else {
				dst.Set(sx, sy, 'o') // Body
			}
		}
	}
}

// renderOverlay draws a centered overlay message.
func (g *Game) renderOverlay(dst *core.Screen, line1, line2 string) {
	w := dst.Width()
	h := dst.Height()

	// Calculate box dimensions
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

// drawCenteredText draws text centered horizontally.
func (g *Game) drawCenteredText(dst *core.Screen, text string, y int) {
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
func (g *Game) State() core.GameState {
	return core.GameState{
		Score:    g.score,
		GameOver: g.gameOver || g.won,
		Paused:   g.paused,
	}
}

// --- String representation for Direction ---

func (d Direction) String() string {
	switch d {
	case DirUp:
		return "up"
	case DirDown:
		return "down"
	case DirLeft:
		return "left"
	case DirRight:
		return "right"
	default:
		return "unknown"
	}
}

// --- Helper for level names (used in UI) ---

// LevelNames returns the names of all levels.
func LevelNames() []string {
	names := make([]string, LevelCount())
	for i, level := range Levels {
		names[i] = level.Name
	}
	return names
}

// --- Debug helper ---

// DebugState returns a string representation of the game state.
func (g *Game) DebugState() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Tick: %d, Score: %d, Level: %d\n", g.tick, g.score, g.levelIndex+1))
	b.WriteString(fmt.Sprintf("Snake len: %d, Direction: %s\n", len(g.snake), g.direction))
	if len(g.snake) > 0 {
		b.WriteString(fmt.Sprintf("Head: (%d, %d), Food: (%d, %d)\n", g.snake[0].X, g.snake[0].Y, g.food.X, g.food.Y))
	}
	b.WriteString(fmt.Sprintf("GameOver: %v, Won: %v, Paused: %v\n", g.gameOver, g.won, g.paused))
	return b.String()
}
