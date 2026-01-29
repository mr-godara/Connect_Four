package game

import "fmt"

const (
	Rows = 6
	Cols = 7
)

type Board struct {
	Grid [][]int
}

func NewBoard() *Board {
	grid := make([][]int, Rows)
	for i := range grid {
		grid[i] = make([]int, Cols)
	}
	return &Board{Grid: grid}
}

// DropDisc drops a disc in a column, returns row index or -1 if invalid
func (b *Board) DropDisc(col, player int) int {
	if col < 0 || col >= Cols {
		return -1
	}

	// Find the lowest empty row in the column
	for row := Rows - 1; row >= 0; row-- {
		if b.Grid[row][col] == 0 {
			b.Grid[row][col] = player
			return row
		}
	}
	return -1 // Column is full
}

// IsValidMove checks if a move is valid
func (b *Board) IsValidMove(col int) bool {
	if col < 0 || col >= Cols {
		return false
	}
	return b.Grid[0][col] == 0
}

// GetValidMoves returns all valid column indices
func (b *Board) GetValidMoves() []int {
	validMoves := []int{}
	for col := 0; col < Cols; col++ {
		if b.IsValidMove(col) {
			validMoves = append(validMoves, col)
		}
	}
	return validMoves
}

type WinResult struct {
	HasWinner bool
	Positions [][2]int
}

// CheckWinner checks for a winner after a move at (row, col)
func (b *Board) CheckWinner(row, col int) WinResult {
	player := b.Grid[row][col]
	if player == 0 {
		return WinResult{HasWinner: false}
	}

	// Check all four directions
	directions := [][2]int{
		{0, 1},  // Horizontal
		{1, 0},  // Vertical
		{1, 1},  // Diagonal down-right
		{1, -1}, // Diagonal down-left
	}

	for _, dir := range directions {
		result := b.checkDirection(row, col, dir[0], dir[1], player)
		if result.count >= 4 {
			return WinResult{HasWinner: true, Positions: result.positions}
		}
	}

	return WinResult{HasWinner: false}
}

type directionResult struct {
	count     int
	positions [][2]int
}

func (b *Board) checkDirection(row, col, dRow, dCol, player int) directionResult {
	count := 1
	positions := [][2]int{{row, col}}

	// Check forward direction
	r, c := row+dRow, col+dCol
	for b.isInBounds(r, c) && b.Grid[r][c] == player {
		count++
		positions = append(positions, [2]int{r, c})
		r += dRow
		c += dCol
	}

	// Check backward direction
	r, c = row-dRow, col-dCol
	for b.isInBounds(r, c) && b.Grid[r][c] == player {
		count++
		positions = append([][2]int{{r, c}}, positions...)
		r -= dRow
		c -= dCol
	}

	return directionResult{count: count, positions: positions}
}

func (b *Board) isInBounds(row, col int) bool {
	return row >= 0 && row < Rows && col >= 0 && col < Cols
}

// IsFull checks if the board is full (draw condition)
func (b *Board) IsFull() bool {
	for col := 0; col < Cols; col++ {
		if b.Grid[0][col] == 0 {
			return false
		}
	}
	return true
}

// Clone creates a deep copy of the board
func (b *Board) Clone() *Board {
	clone := NewBoard()
	for i := range b.Grid {
		copy(clone.Grid[i], b.Grid[i])
	}
	return clone
}

// GetState returns a copy of the grid
func (b *Board) GetState() [][]int {
	state := make([][]int, Rows)
	for i := range b.Grid {
		state[i] = make([]int, Cols)
		copy(state[i], b.Grid[i])
	}
	return state
}

func (b *Board) Print() {
	for _, row := range b.Grid {
		for _, cell := range row {
			fmt.Printf("%d ", cell)
		}
		fmt.Println()
	}
}
