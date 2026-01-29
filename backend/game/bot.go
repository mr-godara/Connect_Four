package game

type Bot struct {
	PlayerNumber   int
	OpponentNumber int
}

func NewBot(playerNumber int) *Bot {
	opponentNumber := 1
	if playerNumber == 1 {
		opponentNumber = 2
	}
	return &Bot{
		PlayerNumber:   playerNumber,
		OpponentNumber: opponentNumber,
	}
}

// GetBestMove returns the best column for the bot to play
func (bot *Bot) GetBestMove(board *Board) int {
	validMoves := board.GetValidMoves()
	if len(validMoves) == 0 {
		return -1
	}

	// Priority 1: Check if bot can win
	winningMove := bot.findWinningMove(board, bot.PlayerNumber)
	if winningMove != -1 {
		return winningMove
	}

	// Priority 2: Block opponent's winning move
	blockingMove := bot.findWinningMove(board, bot.OpponentNumber)
	if blockingMove != -1 {
		return blockingMove
	}

	// Priority 3: Prefer center columns
	centerPreference := []int{3, 2, 4, 1, 5, 0, 6}
	for _, col := range centerPreference {
		if board.IsValidMove(col) {
			if !bot.createsOpponentWin(board, col) {
				return col
			}
		}
	}

	// Priority 4: Play any valid move that doesn't give opponent immediate win
	for _, col := range validMoves {
		if !bot.createsOpponentWin(board, col) {
			return col
		}
	}

	// If all moves create opponent win opportunity, pick the best available
	return validMoves[0]
}

func (bot *Bot) findWinningMove(board *Board, player int) int {
	validMoves := board.GetValidMoves()

	for _, col := range validMoves {
		clonedBoard := board.Clone()
		row := clonedBoard.DropDisc(col, player)

		if row != -1 {
			winResult := clonedBoard.CheckWinner(row, col)
			if winResult.HasWinner {
				return col
			}
		}
	}

	return -1
}

func (bot *Bot) createsOpponentWin(board *Board, col int) bool {
	clonedBoard := board.Clone()
	row := clonedBoard.DropDisc(col, bot.PlayerNumber)

	if row == -1 {
		return false
	}

	// Check if opponent can win after this move
	opponentWinningMove := bot.findWinningMove(clonedBoard, bot.OpponentNumber)
	return opponentWinningMove != -1
}
