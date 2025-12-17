package breakout

// Fixed-point scale factor: 1 cell = 1000 units.
// This allows for sub-pixel precision while maintaining determinism.
const Scale = 1000

// Fixed represents a fixed-point integer (scaled by Scale).
type Fixed int

// ToFixed converts a cell coordinate to fixed-point.
func ToFixed(cell int) Fixed {
	return Fixed(cell * Scale)
}

// ToCell converts fixed-point to cell coordinate (truncated).
func (f Fixed) ToCell() int {
	return int(f) / Scale
}

// ToCellRounded converts fixed-point to nearest cell.
func (f Fixed) ToCellRounded() int {
	if f >= 0 {
		return int(f+Scale/2) / Scale
	}
	return int(f-Scale/2) / Scale
}

// Add adds two fixed-point values.
func (f Fixed) Add(other Fixed) Fixed {
	return f + other
}

// Sub subtracts two fixed-point values.
func (f Fixed) Sub(other Fixed) Fixed {
	return f - other
}

// Mul multiplies fixed-point by an integer.
func (f Fixed) Mul(n int) Fixed {
	return Fixed(int(f) * n)
}

// Div divides fixed-point by an integer.
func (f Fixed) Div(n int) Fixed {
	if n == 0 {
		return 0
	}
	return Fixed(int(f) / n)
}

// Abs returns absolute value.
func (f Fixed) Abs() Fixed {
	if f < 0 {
		return -f
	}
	return f
}

// Sign returns -1, 0, or 1.
func (f Fixed) Sign() int {
	if f < 0 {
		return -1
	}
	if f > 0 {
		return 1
	}
	return 0
}

// Ball represents the ball state with fixed-point coordinates.
type Ball struct {
	X, Y   Fixed // Position (center)
	VX, VY Fixed // Velocity per tick
	Stuck  bool  // Whether ball is stuck to paddle (sticky power-up)
	Active bool  // Whether ball is in play (for multi-ball)
}

// CellX returns the ball's X position in cell coordinates.
func (b *Ball) CellX() int {
	return b.X.ToCell()
}

// CellY returns the ball's Y position in cell coordinates.
func (b *Ball) CellY() int {
	return b.Y.ToCell()
}

// Move updates ball position by velocity.
func (b *Ball) Move() {
	b.X = b.X.Add(b.VX)
	b.Y = b.Y.Add(b.VY)
}

// BounceX reverses horizontal velocity.
func (b *Ball) BounceX() {
	b.VX = -b.VX
}

// BounceY reverses vertical velocity.
func (b *Ball) BounceY() {
	b.VY = -b.VY
}

// Paddle represents the player's paddle.
type Paddle struct {
	X     Fixed // Left edge position (fixed-point)
	Y     int   // Cell Y position (fixed row at bottom)
	Width int   // Width in cells
}

// CellX returns paddle's left edge in cell coordinates.
func (p *Paddle) CellX() int {
	return p.X.ToCell()
}

// CenterX returns paddle's center in fixed-point.
func (p *Paddle) CenterX() Fixed {
	return p.X.Add(ToFixed(p.Width).Div(2))
}

// Left returns left edge in fixed-point.
func (p *Paddle) Left() Fixed {
	return p.X
}

// Right returns right edge in fixed-point.
func (p *Paddle) Right() Fixed {
	return p.X.Add(ToFixed(p.Width))
}

// CollisionSide indicates which side of an object was hit.
type CollisionSide int

const (
	CollisionNone CollisionSide = iota
	CollisionTop
	CollisionBottom
	CollisionLeft
	CollisionRight
)

// CheckWallCollision checks if ball hits screen boundaries.
// Returns the collision side and whether the ball fell below the screen.
func CheckWallCollision(ball *Ball, screenW, screenH int) (side CollisionSide, fellOff bool) {
	// Left wall
	if ball.X < ToFixed(1) {
		ball.X = ToFixed(1)
		return CollisionLeft, false
	}

	// Right wall
	if ball.X >= ToFixed(screenW-1) {
		ball.X = ToFixed(screenW - 2)
		return CollisionRight, false
	}

	// Top wall (HUD area usually at row 0)
	if ball.Y < ToFixed(2) {
		ball.Y = ToFixed(2)
		return CollisionTop, false
	}

	// Bottom - ball fell off
	if ball.Y >= ToFixed(screenH) {
		return CollisionBottom, true
	}

	return CollisionNone, false
}

// CheckPaddleCollision checks if ball hits the paddle.
// If collision occurs, adjusts ball velocity based on hit position.
// Returns true if collision occurred.
func CheckPaddleCollision(ball *Ball, paddle *Paddle, baseSpeed Fixed) bool {
	// Ball must be moving downward and at paddle's Y level
	if ball.VY <= 0 {
		return false
	}

	ballY := ball.Y.ToCell()
	if ballY != paddle.Y && ballY != paddle.Y-1 {
		return false
	}

	// Check horizontal overlap
	ballX := ball.X
	paddleLeft := paddle.Left()
	paddleRight := paddle.Right()

	if ballX < paddleLeft || ballX > paddleRight {
		return false
	}

	// Collision! Calculate where on paddle the ball hit
	// Range: -1.0 (left edge) to +1.0 (right edge)
	paddleCenter := paddle.CenterX()
	hitOffset := ballX.Sub(paddleCenter)
	halfWidth := ToFixed(paddle.Width).Div(2)

	// Normalize to -1000 to +1000 range
	var normalizedHit Fixed
	if halfWidth > 0 {
		normalizedHit = hitOffset.Mul(Scale).Div(int(halfWidth))
	}

	// Bounce upward
	ball.VY = -ball.VY.Abs()
	if ball.VY > -baseSpeed/2 {
		ball.VY = -baseSpeed / 2
	}

	// Adjust horizontal velocity based on hit position
	// Edge hits give more horizontal angle
	ball.VX = normalizedHit.Mul(int(baseSpeed)) / Scale

	// Ensure ball moves away from paddle
	ball.Y = ToFixed(paddle.Y - 1)

	return true
}

// CheckBrickCollision checks if ball hits any brick in the level.
// Returns the brick coordinates (row, col) and collision side, or (-1, -1, CollisionNone) if no hit.
func CheckBrickCollision(ball *Ball, level *Level, brickAreaTop, brickHeight, brickWidth int) (row, col int, side CollisionSide) {
	ballCellX := ball.CellX()
	ballCellY := ball.CellY()

	// Calculate which brick cell the ball is in
	row = (ballCellY - brickAreaTop) / brickHeight
	col = ballCellX / brickWidth

	// Check bounds
	if row < 0 || row >= level.Height || col < 0 || col >= level.Width {
		return -1, -1, CollisionNone
	}

	brick := &level.Bricks[row][col]
	if !brick.Alive || brick.Type == BrickEmpty {
		return -1, -1, CollisionNone
	}

	// Brick boundaries
	brickLeft := col * brickWidth
	brickRight := brickLeft + brickWidth
	brickTop := brickAreaTop + row*brickHeight
	brickBottom := brickTop + brickHeight

	// Determine collision side based on ball velocity and position
	// Simple approach: check which edge the ball is closest to

	distLeft := ball.X.Sub(ToFixed(brickLeft)).Abs()
	distRight := ball.X.Sub(ToFixed(brickRight)).Abs()
	distTop := ball.Y.Sub(ToFixed(brickTop)).Abs()
	distBottom := ball.Y.Sub(ToFixed(brickBottom)).Abs()

	minHoriz := distLeft
	horizSide := CollisionLeft
	if distRight < minHoriz {
		minHoriz = distRight
		horizSide = CollisionRight
	}

	minVert := distTop
	vertSide := CollisionTop
	if distBottom < minVert {
		minVert = distBottom
		vertSide = CollisionBottom
	}

	// Prefer vertical bounce if ball is moving mostly vertically
	if ball.VY.Abs() > ball.VX.Abs() || minVert <= minHoriz {
		return row, col, vertSide
	}
	return row, col, horizSide
}

// ApplyCollisionBounce applies the appropriate bounce based on collision side.
func ApplyCollisionBounce(ball *Ball, side CollisionSide) {
	switch side {
	case CollisionTop, CollisionBottom:
		ball.BounceY()
	case CollisionLeft, CollisionRight:
		ball.BounceX()
	}
}

// Clamp restricts a value to [minVal, maxVal].
func ClampFixed(val, minVal, maxVal Fixed) Fixed {
	if val < minVal {
		return minVal
	}
	if val > maxVal {
		return maxVal
	}
	return val
}
