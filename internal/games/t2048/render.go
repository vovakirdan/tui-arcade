package t2048

import (
	"fmt"
	"strconv"

	"github.com/vovakirdan/tui-arcade/internal/core"
)

// Render draws the game state to the screen.
func (g *Game) Render(dst *core.Screen) {
	dst.Clear()

	// Check screen size
	if g.tooSmall {
		g.renderTooSmall(dst)
		return
	}

	// Calculate board position (centered)
	boardW := BoardSize*g.cellWidth + 1  // +1 for right border
	boardH := BoardSize*g.cellHeight + 1 // +1 for bottom border

	boardX := (g.screenW - boardW) / 2
	boardY := hudHeight + 1

	// Render HUD
	g.renderHUD(dst, boardX, boardW)

	// Render board grid
	g.renderBoardGrid(dst, boardX, boardY)

	// Render tiles (animated or static)
	if g.animating {
		g.renderAnimatedTiles(dst, boardX, boardY)
	} else {
		g.renderStaticTiles(dst, boardX, boardY)
	}

	// Render overlays
	g.renderOverlays(dst, boardX, boardY, boardW, boardH)
}

// renderTooSmall shows a "window too small" message.
func (g *Game) renderTooSmall(dst *core.Screen) {
	msg := "Window too small"
	x := (g.screenW - len(msg)) / 2
	y := g.screenH / 2
	dst.DrawText(x, y, msg)

	hint := "Please resize terminal"
	hintX := (g.screenW - len(hint)) / 2
	dst.DrawText(hintX, y+1, hint)
}

// renderHUD draws the score and level info.
func (g *Game) renderHUD(dst *core.Screen, boardX, boardW int) {
	// Title with bright yellow
	title := "2048"
	titleX := boardX + (boardW-len(title))/2
	dst.DrawTextWithColor(titleX, 0, title, core.ColorBrightYellow)

	// Score with cyan
	scoreStr := fmt.Sprintf("Score: %d", g.score)
	dst.DrawTextWithColor(boardX, 1, scoreStr, core.ColorCyan)

	// Level/Target info (campaign) or Max tile (endless)
	var infoStr string
	if g.mode == ModeCampaign {
		infoStr = fmt.Sprintf("Level %d/%d  Target: %d", g.levelIndex+1, LevelCount(), g.currentTarget)
	} else {
		infoStr = fmt.Sprintf("Max: %d", MaxTile(g.board))
	}

	infoX := boardX + boardW - len(infoStr)
	if infoX < boardX {
		infoX = boardX
	}
	dst.DrawTextWithColor(infoX, 1, infoStr, core.ColorCyan)

	// Mode indicator with gray
	modeStr := "Campaign"
	if g.mode == ModeEndless {
		modeStr = "Endless"
	}
	modeX := boardX + (boardW-len(modeStr))/2
	dst.DrawTextWithColor(modeX, 2, modeStr, core.ColorGray)
}

// renderBoardGrid draws the 4x4 grid borders (without tiles).
func (g *Game) renderBoardGrid(dst *core.Screen, boardX, boardY int) {
	// Draw grid borders with gray color
	for y := range BoardSize + 1 {
		for x := range BoardSize + 1 {
			px := boardX + x*g.cellWidth
			py := boardY + y*g.cellHeight

			// Draw corner/intersection
			var corner rune
			switch {
			case y == 0 && x == 0:
				corner = '┌'
			case y == 0 && x == BoardSize:
				corner = '┐'
			case y == BoardSize && x == 0:
				corner = '└'
			case y == BoardSize && x == BoardSize:
				corner = '┘'
			case y == 0:
				corner = '┬'
			case y == BoardSize:
				corner = '┴'
			case x == 0:
				corner = '├'
			case x == BoardSize:
				corner = '┤'
			default:
				corner = '┼'
			}
			dst.SetWithColor(px, py, corner, core.ColorGray)

			// Draw horizontal line to the right
			if x < BoardSize {
				for i := 1; i < g.cellWidth; i++ {
					dst.SetWithColor(px+i, py, '─', core.ColorGray)
				}
			}

			// Draw vertical line down
			if y < BoardSize {
				for i := 1; i < g.cellHeight; i++ {
					dst.SetWithColor(px, py+i, '│', core.ColorGray)
				}
			}
		}
	}
}

// renderStaticTiles draws tiles at their logical positions.
func (g *Game) renderStaticTiles(dst *core.Screen, boardX, boardY int) {
	for y := range BoardSize {
		for x := range BoardSize {
			val := g.board[y][x]
			if val == 0 {
				continue
			}
			g.renderTileAt(dst, boardX, boardY, x, y, val)
		}
	}
}

// renderAnimatedTiles draws tiles at interpolated positions during animation.
func (g *Game) renderAnimatedTiles(dst *core.Screen, boardX, boardY int) {
	// Track which logical positions are being animated (to avoid drawing static tiles there)
	animatedFrom := make(map[[2]int]bool)
	animatedTo := make(map[[2]int]bool)

	for _, anim := range g.animations {
		animatedFrom[[2]int{anim.FromX, anim.FromY}] = true
		animatedTo[[2]int{anim.ToX, anim.ToY}] = true
	}

	// Draw static tiles that are NOT involved in animation
	for y := range BoardSize {
		for x := range BoardSize {
			val := g.board[y][x]
			if val == 0 {
				continue
			}
			// Skip if this position is a destination of an animation (we'll draw it animated)
			if animatedTo[[2]int{x, y}] {
				continue
			}
			g.renderTileAt(dst, boardX, boardY, x, y, val)
		}
	}

	// Draw animated tiles
	for _, anim := range g.animations {
		if anim.IsNew {
			// Pop animation for new tiles
			if anim.Progress >= 0.3 {
				// Tile appears after 30% of pop animation
				g.renderTileAt(dst, boardX, boardY, anim.ToX, anim.ToY, anim.Value)
			}
		} else {
			// Slide animation
			interpX, interpY := anim.interpolatePosition()
			g.renderTileAtFloat(dst, boardX, boardY, interpX, interpY, anim.Value)
		}
	}
}

// renderTileAt draws a tile value at a logical grid position.
func (g *Game) renderTileAt(dst *core.Screen, boardX, boardY, cellX, cellY, val int) {
	// Calculate pixel position for center of cell
	px := boardX + cellX*g.cellWidth + 1
	py := boardY + cellY*g.cellHeight + g.cellHeight/2

	// Format and center value
	valStr := strconv.Itoa(val)
	padLeft := (g.cellWidth - 1 - len(valStr)) / 2
	if padLeft < 0 {
		padLeft = 0
	}

	dst.DrawTextWithColor(px+padLeft, py, valStr, tileColor(val))
}

// renderTileAtFloat draws a tile value at interpolated float position.
func (g *Game) renderTileAtFloat(dst *core.Screen, boardX, boardY int, cellX, cellY float64, val int) {
	// Calculate pixel position with float interpolation
	px := boardX + int(cellX*float64(g.cellWidth)+0.5) + 1
	py := boardY + int(cellY*float64(g.cellHeight)+0.5) + g.cellHeight/2

	// Format and center value
	valStr := strconv.Itoa(val)
	padLeft := (g.cellWidth - 1 - len(valStr)) / 2
	if padLeft < 0 {
		padLeft = 0
	}

	dst.DrawTextWithColor(px+padLeft, py, valStr, tileColor(val))
}

// renderOverlays draws game state overlays.
func (g *Game) renderOverlays(dst *core.Screen, boardX, boardY, boardW, boardH int) {
	centerX := boardX + boardW/2
	centerY := boardY + boardH/2

	if g.paused {
		g.drawOverlay(dst, centerX, centerY, "PAUSED", "Press P to resume")
		return
	}

	if g.levelCleared {
		targetStr := fmt.Sprintf("Target %d reached!", g.currentTarget)
		if g.levelIndex >= LevelCount()-1 {
			g.drawOverlay(dst, centerX, centerY, targetStr, "Final level complete!")
		} else {
			nextStr := fmt.Sprintf("Next: Level %d", g.levelIndex+2)
			g.drawOverlay(dst, centerX, centerY, targetStr, nextStr)
		}
		return
	}

	if g.won {
		g.drawOverlay(dst, centerX, centerY, "CAMPAIGN COMPLETE!", "You are the champion!", "Press R to restart")
		return
	}

	if g.gameOver {
		maxStr := fmt.Sprintf("Max tile: %d", MaxTile(g.board))
		g.drawOverlay(dst, centerX, centerY, "GAME OVER", maxStr, "Press R to restart")
		return
	}
}

// drawOverlay draws a centered text overlay.
func (g *Game) drawOverlay(dst *core.Screen, centerX, centerY int, lines ...string) {
	// Find max line width
	maxLen := 0
	for _, line := range lines {
		if len(line) > maxLen {
			maxLen = len(line)
		}
	}

	// Draw box
	boxW := maxLen + 4
	boxH := len(lines) + 2
	boxX := centerX - boxW/2
	boxY := centerY - boxH/2

	// Clear area behind overlay
	for y := boxY; y < boxY+boxH; y++ {
		for x := boxX; x < boxX+boxW; x++ {
			dst.Set(x, y, ' ')
		}
	}

	// Draw border
	dst.DrawBox(core.Rect{X: boxX, Y: boxY, W: boxW, H: boxH})

	// Draw text
	for i, line := range lines {
		x := centerX - len(line)/2
		dst.DrawText(x, boxY+1+i, line)
	}
}

// Controls returns the control hints for the game.
func (g *Game) Controls() string {
	return "Arrow keys/WASD: Move | P: Pause | R: Restart | Q: Quit"
}

// tileColor returns the color for a tile based on its value.
func tileColor(val int) core.Color {
	switch {
	case val <= 4:
		return core.ColorWhite
	case val <= 16:
		return core.ColorYellow
	case val <= 64:
		return core.ColorOrange
	case val <= 256:
		return core.ColorRed
	case val <= 1024:
		return core.ColorMagenta
	default:
		return core.ColorBrightMagenta
	}
}
