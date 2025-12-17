package t2048

// Direction represents a move direction.
type Direction int

const (
	DirUp Direction = iota
	DirDown
	DirLeft
	DirRight
)

// BoardSize is the default board dimension.
const BoardSize = 4

// Board represents a 4x4 game board.
type Board [BoardSize][BoardSize]int

// slideRow slides and merges a single row to the left.
// Returns the updated row and the score gained from merges.
func slideRow(row [BoardSize]int) (result [BoardSize]int, score int) {
	writePos := 0

	for i := range BoardSize {
		if row[i] == 0 {
			continue
		}

		if writePos > 0 && result[writePos-1] == row[i] {
			// Merge with previous tile
			result[writePos-1] *= 2
			score += result[writePos-1]
		} else {
			// Move tile
			result[writePos] = row[i]
			writePos++
		}
	}

	return result, score
}

// reverseRow reverses a row.
func reverseRow(row [BoardSize]int) [BoardSize]int {
	var result [BoardSize]int
	for i := range BoardSize {
		result[i] = row[BoardSize-1-i]
	}
	return result
}

// SlideLeft slides all tiles left and merges.
// Returns the new board, score gained, and whether the board changed.
func SlideLeft(board Board) (Board, int, bool) {
	var newBoard Board
	totalScore := 0
	changed := false

	for y := range BoardSize {
		row := board[y]
		newRow, score := slideRow(row)
		newBoard[y] = newRow
		totalScore += score

		if row != newRow {
			changed = true
		}
	}

	return newBoard, totalScore, changed
}

// SlideRight slides all tiles right and merges.
func SlideRight(board Board) (Board, int, bool) {
	var newBoard Board
	totalScore := 0
	changed := false

	for y := range BoardSize {
		// Reverse, slide left, reverse back
		row := reverseRow(board[y])
		newRow, score := slideRow(row)
		newBoard[y] = reverseRow(newRow)
		totalScore += score

		if board[y] != newBoard[y] {
			changed = true
		}
	}

	return newBoard, totalScore, changed
}

// SlideUp slides all tiles up and merges.
func SlideUp(board Board) (Board, int, bool) {
	// Transpose, slide left, transpose back
	transposed := transpose(board)
	slid, score, changed := SlideLeft(transposed)
	return transpose(slid), score, changed
}

// SlideDown slides all tiles down and merges.
func SlideDown(board Board) (Board, int, bool) {
	// Transpose, slide right, transpose back
	transposed := transpose(board)
	slid, score, changed := SlideRight(transposed)
	return transpose(slid), score, changed
}

// transpose returns the matrix transpose.
func transpose(board Board) Board {
	var result Board
	for y := range BoardSize {
		for x := range BoardSize {
			result[y][x] = board[x][y]
		}
	}
	return result
}

// Slide performs a move in the given direction.
// Returns the new board, score gained, and whether the board changed.
func Slide(board Board, dir Direction) (Board, int, bool) {
	switch dir {
	case DirLeft:
		return SlideLeft(board)
	case DirRight:
		return SlideRight(board)
	case DirUp:
		return SlideUp(board)
	case DirDown:
		return SlideDown(board)
	default:
		return board, 0, false
	}
}

// EmptyCells returns coordinates of all empty cells.
func EmptyCells(board Board) []struct{ X, Y int } {
	var cells []struct{ X, Y int }
	for y := range BoardSize {
		for x := range BoardSize {
			if board[y][x] == 0 {
				cells = append(cells, struct{ X, Y int }{x, y})
			}
		}
	}
	return cells
}

// HasEmptyCell returns true if there's at least one empty cell.
func HasEmptyCell(board Board) bool {
	for y := range BoardSize {
		for x := range BoardSize {
			if board[y][x] == 0 {
				return true
			}
		}
	}
	return false
}

// HasPossibleMerge returns true if any adjacent tiles can merge.
func HasPossibleMerge(board Board) bool {
	for y := range BoardSize {
		for x := range BoardSize {
			val := board[y][x]
			// Check right neighbor
			if x < BoardSize-1 && board[y][x+1] == val {
				return true
			}
			// Check bottom neighbor
			if y < BoardSize-1 && board[y+1][x] == val {
				return true
			}
		}
	}
	return false
}

// CanMove returns true if any move is possible.
func CanMove(board Board) bool {
	return HasEmptyCell(board) || HasPossibleMerge(board)
}

// MaxTile returns the maximum tile value on the board.
func MaxTile(board Board) int {
	maxVal := 0
	for y := range BoardSize {
		for x := range BoardSize {
			if board[y][x] > maxVal {
				maxVal = board[y][x]
			}
		}
	}
	return maxVal
}

// IsGameOver returns true if no moves are possible.
func IsGameOver(board Board) bool {
	return !CanMove(board)
}

// SlideWithTracking performs a move and returns tile movement tracking for animation.
func SlideWithTracking(board Board, dir Direction) (newBoard Board, moves []TileMove, score int, changed bool) {
	switch dir {
	case DirLeft:
		return slideLeftWithTracking(board)
	case DirRight:
		return slideRightWithTracking(board)
	case DirUp:
		return slideUpWithTracking(board)
	case DirDown:
		return slideDownWithTracking(board)
	default:
		return board, nil, 0, false
	}
}

// slideRowWithTracking slides a row left and returns move tracking.
func slideRowWithTracking(row [BoardSize]int, y int) (result [BoardSize]int, moves []TileMove, score int) {
	writePos := 0

	for x := range BoardSize {
		if row[x] == 0 {
			continue
		}

		if writePos > 0 && result[writePos-1] == row[x] {
			// Merge with previous tile
			result[writePos-1] *= 2
			score += result[writePos-1]
			moves = append(moves, TileMove{
				FromX:  x,
				FromY:  y,
				ToX:    writePos - 1,
				ToY:    y,
				Value:  row[x],
				Merged: true,
			})
		} else {
			// Move tile
			if x != writePos {
				moves = append(moves, TileMove{
					FromX:  x,
					FromY:  y,
					ToX:    writePos,
					ToY:    y,
					Value:  row[x],
					Merged: false,
				})
			}
			result[writePos] = row[x]
			writePos++
		}
	}

	return result, moves, score
}

// slideLeftWithTracking slides all tiles left with tracking.
func slideLeftWithTracking(board Board) (Board, []TileMove, int, bool) {
	var newBoard Board
	var allMoves []TileMove
	totalScore := 0
	changed := false

	for y := range BoardSize {
		newRow, rowMoves, score := slideRowWithTracking(board[y], y)
		newBoard[y] = newRow
		allMoves = append(allMoves, rowMoves...)
		totalScore += score

		if board[y] != newRow {
			changed = true
		}
	}

	return newBoard, allMoves, totalScore, changed
}

// slideRightWithTracking slides all tiles right with tracking.
func slideRightWithTracking(board Board) (Board, []TileMove, int, bool) {
	var newBoard Board
	var allMoves []TileMove
	totalScore := 0
	changed := false

	for y := range BoardSize {
		// Reverse row, slide left, reverse back
		reversed := reverseRow(board[y])
		slidRow, rowMoves, score := slideRowWithTracking(reversed, y)
		newBoard[y] = reverseRow(slidRow)
		totalScore += score

		// Fix move coordinates (mirror X positions)
		for i := range rowMoves {
			rowMoves[i].FromX = BoardSize - 1 - rowMoves[i].FromX
			rowMoves[i].ToX = BoardSize - 1 - rowMoves[i].ToX
		}
		allMoves = append(allMoves, rowMoves...)

		if board[y] != newBoard[y] {
			changed = true
		}
	}

	return newBoard, allMoves, totalScore, changed
}

// slideUpWithTracking slides all tiles up with tracking.
func slideUpWithTracking(board Board) (Board, []TileMove, int, bool) {
	// Transpose, slide left, transpose back
	transposed := transpose(board)
	slidBoard, moves, score, changed := slideLeftWithTracking(transposed)

	// Fix move coordinates (swap X/Y)
	for i := range moves {
		moves[i].FromX, moves[i].FromY = moves[i].FromY, moves[i].FromX
		moves[i].ToX, moves[i].ToY = moves[i].ToY, moves[i].ToX
	}

	return transpose(slidBoard), moves, score, changed
}

// slideDownWithTracking slides all tiles down with tracking.
func slideDownWithTracking(board Board) (Board, []TileMove, int, bool) {
	// Transpose, slide right, transpose back
	transposed := transpose(board)
	slidBoard, moves, score, changed := slideRightWithTracking(transposed)

	// Fix move coordinates (swap X/Y)
	for i := range moves {
		moves[i].FromX, moves[i].FromY = moves[i].FromY, moves[i].FromX
		moves[i].ToX, moves[i].ToY = moves[i].ToY, moves[i].ToX
	}

	return transpose(slidBoard), moves, score, changed
}
