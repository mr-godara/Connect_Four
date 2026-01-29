package game

import (
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Player struct {
	Username string
	Conn     *websocket.Conn
	IsBot    bool
}

type Game struct {
	ID                   string
	Player1              *Player
	Player2              *Player
	Board                *Board
	CurrentTurn          int
	Status               string
	Winner               string
	IsBot                bool
	Bot                  *Bot
	StartTime            time.Time
	LastActivityTime     time.Time
	DisconnectedPlayers  map[string]bool
}

type MoveResult struct {
	Success          bool
	Error            string
	Row              int
	Col              int
	GameOver         bool
	Winner           string
	WinningPositions [][2]int
	NextTurn         int
	Board            [][]int
}

func NewGame(player1, player2 *Player, isBot bool) *Game {
	game := &Game{
		ID:                  uuid.New().String(),
		Player1:             player1,
		Player2:             player2,
		Board:               NewBoard(),
		CurrentTurn:         1,
		Status:              "active",
		IsBot:               isBot,
		StartTime:           time.Now(),
		LastActivityTime:    time.Now(),
		DisconnectedPlayers: make(map[string]bool),
	}

	if isBot {
		game.Bot = NewBot(2)
	}

	return game
}

func (g *Game) MakeMove(playerNumber, col int) MoveResult {
	// Validate it's the player's turn
	if g.CurrentTurn != playerNumber {
		return MoveResult{Success: false, Error: "Not your turn"}
	}

	// Validate the move
	if !g.Board.IsValidMove(col) {
		return MoveResult{Success: false, Error: "Invalid move"}
	}

	// Make the move
	row := g.Board.DropDisc(col, playerNumber)
	if row == -1 {
		return MoveResult{Success: false, Error: "Invalid move"}
	}

	g.LastActivityTime = time.Now()

	// Check for winner
	winResult := g.Board.CheckWinner(row, col)

	if winResult.HasWinner {
		g.Status = "completed"
		if playerNumber == 1 {
			g.Winner = g.Player1.Username
		} else {
			g.Winner = g.Player2.Username
		}

		return MoveResult{
			Success:          true,
			Row:              row,
			Col:              col,
			GameOver:         true,
			Winner:           g.Winner,
			WinningPositions: winResult.Positions,
			Board:            g.Board.GetState(),
		}
	}

	// Check for draw
	if g.Board.IsFull() {
		g.Status = "completed"
		g.Winner = "draw"

		return MoveResult{
			Success:  true,
			Row:      row,
			Col:      col,
			GameOver: true,
			Winner:   "draw",
			Board:    g.Board.GetState(),
		}
	}

	// Switch turn
	if g.CurrentTurn == 1 {
		g.CurrentTurn = 2
	} else {
		g.CurrentTurn = 1
	}

	return MoveResult{
		Success:  true,
		Row:      row,
		Col:      col,
		GameOver: false,
		NextTurn: g.CurrentTurn,
		Board:    g.Board.GetState(),
	}
}

func (g *Game) MakeBotMove() *MoveResult {
	if g.Bot == nil || g.Status != "active" {
		return nil
	}

	col := g.Bot.GetBestMove(g.Board)
	if col == -1 {
		return nil
	}

	result := g.MakeMove(2, col)
	if !result.Success {
		return nil
	}

	return &result
}

func (g *Game) GetGameState() map[string]interface{} {
	return map[string]interface{}{
		"id":          g.ID,
		"player1":     g.Player1.Username,
		"player2":     g.Player2.Username,
		"board":       g.Board.GetState(),
		"currentTurn": g.CurrentTurn,
		"status":      g.Status,
		"winner":      g.Winner,
		"isBot":       g.IsBot,
	}
}

func (g *Game) GetGameDuration() int {
	return int(time.Since(g.StartTime).Seconds())
}
