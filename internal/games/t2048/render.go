package t2048

import (
	"fmt"
	"strconv"

	"github.com/vovakirdan/tui-arcade/internal/core"
)

const (
	cellWidth  = 5 // Width of each cell (including borders)
	cellHeight = 2 // Height of each cell (including borders)
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
	boardW := BoardSize*cellWidth + 1  // +1 for right border
	boardH := BoardSize*cellHeight + 1 // +1 for bottom border
	hudHeight := 3

	boardX := (g.screenW - boardW) / 2
	boardY := hudHeight + 1

	// Render HUD
	g.renderHUD(dst, boardX, boardW)

	// Render board
	g.renderBoard(dst, boardX, boardY)

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
	// Title
	title := "2048"
	titleX := boardX + (boardW-len(title))/2
	dst.DrawText(titleX, 0, title)

	// Score
	scoreStr := fmt.Sprintf("Score: %d", g.score)
	dst.DrawText(boardX, 1, scoreStr)

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
	dst.DrawText(infoX, 1, infoStr)

	// Mode indicator
	modeStr := "Campaign"
	if g.mode == ModeEndless {
		modeStr = "Endless"
	}
	modeX := boardX + (boardW-len(modeStr))/2
	dst.DrawText(modeX, 2, modeStr)
}

// renderBoard draws the 4x4 grid with tiles.
func (g *Game) renderBoard(dst *core.Screen, boardX, boardY int) {
	// Draw grid borders
	for y := range BoardSize + 1 {
		for x := range BoardSize + 1 {
			px := boardX + x*cellWidth
			py := boardY + y*cellHeight

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
			dst.Set(px, py, corner)

			// Draw horizontal line to the right
			if x < BoardSize {
				for i := 1; i < cellWidth; i++ {
					dst.Set(px+i, py, '─')
				}
			}

			// Draw vertical line down
			if y < BoardSize {
				for i := 1; i < cellHeight; i++ {
					dst.Set(px, py+i, '│')
				}
			}
		}
	}

	// Draw tiles
	for y := range BoardSize {
		for x := range BoardSize {
			val := g.board[y][x]
			if val == 0 {
				continue
			}

			// Calculate cell center position
			cellX := boardX + x*cellWidth + 1
			cellY := boardY + y*cellHeight + 1

			// Format value (right-aligned in cell)
			valStr := strconv.Itoa(val)
			// Center the value in the cell
			padLeft := (cellWidth - 1 - len(valStr)) / 2
			if padLeft < 0 {
				padLeft = 0
			}

			dst.DrawText(cellX+padLeft, cellY, valStr)
		}
	}
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
