package manager

import (
	"connect-four/database"
	"connect-four/game"
	"connect-four/kafka"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type MatchmakingPlayer struct {
	Username string
	Conn     *websocket.Conn
	JoinTime time.Time
}

type GameManager struct {
	games              map[string]*game.Game
	playerGames        map[string]string
	matchmakingQueue   []*MatchmakingPlayer
	reconnectionTimers map[string]*time.Timer
	mu                 sync.RWMutex
	db                 *database.DB
}

func NewGameManager(db *database.DB) *GameManager {
	return &GameManager{
		games:              make(map[string]*game.Game),
		playerGames:        make(map[string]string),
		matchmakingQueue:   []*MatchmakingPlayer{},
		reconnectionTimers: make(map[string]*time.Timer),
		db:                 db,
	}
}

type MatchmakingResult struct {
	Success bool
	Error   string
	Game    *game.Game
	Type    string
}

func (gm *GameManager) AddToMatchmaking(username string, conn *websocket.Conn) MatchmakingResult {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	// Check if player is already in a game
	if gameID, exists := gm.playerGames[username]; exists {
		if existingGame, found := gm.games[gameID]; found && existingGame.Status != "completed" {
			return MatchmakingResult{Success: false, Error: "Already in a game"}
		}
		// Clean up completed game
		gm.cleanupGame(gameID)
	}

	// Check if already in queue
	for i, p := range gm.matchmakingQueue {
		if p.Username == username {
			gm.matchmakingQueue[i].Conn = conn
			return MatchmakingResult{Success: false, Error: "Already in matchmaking queue"}
		}
	}

	player := &MatchmakingPlayer{
		Username: username,
		Conn:     conn,
		JoinTime: time.Now(),
	}
	gm.matchmakingQueue = append(gm.matchmakingQueue, player)

	kafka.SendAnalytics("player_joined_matchmaking", map[string]interface{}{"username": username})

	// Try to match with another player
	if len(gm.matchmakingQueue) >= 2 {
		player1 := gm.matchmakingQueue[0]
		player2 := gm.matchmakingQueue[1]
		gm.matchmakingQueue = gm.matchmakingQueue[2:]

		g := gm.createGame(player1, player2, false)
		return MatchmakingResult{Success: true, Game: g, Type: "pvp"}
	}

	// Set timeout for bot game
	go func() {
		time.Sleep(10 * time.Second)
		gm.checkMatchmakingTimeout(username)
	}()

	return MatchmakingResult{Success: true, Type: "waiting"}
}

func (gm *GameManager) checkMatchmakingTimeout(username string) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	// Find player in queue
	playerIndex := -1
	for i, p := range gm.matchmakingQueue {
		if p.Username == username {
			playerIndex = i
			break
		}
	}

	if playerIndex == -1 {
		return // Player already matched
	}

	player := gm.matchmakingQueue[playerIndex]
	gm.matchmakingQueue = append(gm.matchmakingQueue[:playerIndex], gm.matchmakingQueue[playerIndex+1:]...)

	// Create bot game
	botPlayer := &MatchmakingPlayer{
		Username: "BOT",
		Conn:     nil,
	}

	g := gm.createGame(player, botPlayer, true)

	// Notify player
	if player.Conn != nil {
		msg := map[string]interface{}{
			"type":         "gameStart",
			"game":         g.GetGameState(),
			"playerNumber": 1,
			"message":      "No opponent found. Playing against bot!",
		}
		player.Conn.WriteJSON(msg)
	}

	kafka.SendAnalytics("bot_game_started", map[string]interface{}{
		"username": player.Username,
		"gameId":   g.ID,
	})
}

func (gm *GameManager) createGame(player1, player2 *MatchmakingPlayer, isBot bool) *game.Game {
	p1 := &game.Player{Username: player1.Username, Conn: player1.Conn, IsBot: false}
	p2 := &game.Player{Username: player2.Username, Conn: player2.Conn, IsBot: isBot}

	g := game.NewGame(p1, p2, isBot)

	gm.games[g.ID] = g
	gm.playerGames[player1.Username] = g.ID
	if !isBot {
		gm.playerGames[player2.Username] = g.ID
	}

	kafka.SendAnalytics("game_created", map[string]interface{}{
		"gameId":  g.ID,
		"player1": player1.Username,
		"player2": player2.Username,
		"isBot":   isBot,
	})

	// Notify both players
	if p1.Conn != nil {
		msg := map[string]interface{}{
			"type":         "gameStart",
			"game":         g.GetGameState(),
			"playerNumber": 1,
		}
		p1.Conn.WriteJSON(msg)
	}

	if !isBot && p2.Conn != nil {
		msg := map[string]interface{}{
			"type":         "gameStart",
			"game":         g.GetGameState(),
			"playerNumber": 2,
		}
		p2.Conn.WriteJSON(msg)
	}

	return g
}

func (gm *GameManager) MakeMove(username string, col int) game.MoveResult {
	gm.mu.RLock()
	gameID, exists := gm.playerGames[username]
	if !exists {
		gm.mu.RUnlock()
		return game.MoveResult{Success: false, Error: "Not in a game"}
	}

	g, found := gm.games[gameID]
	gm.mu.RUnlock()

	if !found {
		return game.MoveResult{Success: false, Error: "Game not found"}
	}

	playerNumber := 1
	if g.Player2.Username == username {
		playerNumber = 2
	}

	result := g.MakeMove(playerNumber, col)
	if !result.Success {
	kafka.SendAnalytics("move_made", map[string]interface{}{
		"gameId":       gameID,
		"username":     username,
		"playerNumber": playerNumber,
		"col":          col,
		"row":          result.Row,
	})

		return result
	}

	log.Printf("Player %s (%d) placed disc at row %d, col %d", username, playerNumber, result.Row, result.Col)

	// Broadcast move to both players
	gm.broadcastToGame(g, map[string]interface{}{
		"type":        "move",
		"username":    username,
		"playerNumber": playerNumber,
		"col":         result.Col,
		"row":         result.Row,
		"board":       result.Board,
		"nextTurn":    result.NextTurn,
		"gameOver":    result.GameOver,
		"winner":      result.Winner,
		"winningPositions": result.WinningPositions,
	})

	// Handle game over
	if result.GameOver {
		go gm.handleGameOver(g)
	} else if g.IsBot && g.CurrentTurn == 2 {
		// Bot's turn
		go func() {
			time.Sleep(500 * time.Millisecond)
			gm.makeBotMove(g)
		}()
	}

	return result
}

func (gm *GameManager) makeBotMove(g *game.Game) {
	if g == nil || g.Status != "active" {
		return
	}

	result := g.MakeBotMove()
	if result == nil || !result.Success {
		return
	}

	log.Printf("Player BOT (2) placed disc at row %d, col %d", result.Row, result.Col)

	// Broadcast bot move
	gm.broadcastToGame(g, map[string]interface{}{
		"type":             "move",
		"username":         "BOT",
		"playerNumber":     2,
		"col":              result.Col,
		"row":              result.Row,
		"board":            result.Board,
		"nextTurn":         result.NextTurn,
		"gameOver":         result.GameOver,
		"winner":           result.Winner,
		"winningPositions": result.WinningPositions,
	})

	if result.GameOver {
		go gm.handleGameOver(g)
	}
}

func (gm *GameManager) broadcastToGame(g *game.Game, message interface{}) {
	data, _ := json.Marshal(message)

	if g.Player1.Conn != nil {
		g.Player1.Conn.WriteMessage(websocket.TextMessage, data)
	}

	if !g.IsBot && g.Player2.Conn != nil {
		g.Player2.Conn.WriteMessage(websocket.TextMessage, data)
	}
}

func (gm *GameManager) handleGameOver(g *game.Game) {
	duration := g.GetGameDuration()

	// Save to database
	if gm.db != nil {
		if err := gm.db.SaveGame(g.ID, g.Player1.Username, g.Player2.Username, g.Winner, duration, g.IsBot); err != nil {
			log.Printf("Error saving game: %v", err)
		}

		if err := gm.db.UpdateLeaderboard(g); err != nil {
			log.Printf("Error updating leaderboard: %v", err)
		}
	}

	// Send game completion analytics
	kafka.SendAnalytics("game_completed", map[string]interface{}{
		"gameId":   g.ID,
		"player1":  g.Player1.Username,
		"player2":  g.Player2.Username,
		"winner":   g.Winner,
		"duration": duration,
		"isBot":    g.IsBot,
	})

	// Clean up after 30 seconds
	time.AfterFunc(30*time.Second, func() {
		gm.mu.Lock()
		defer gm.mu.Unlock()
		gm.cleanupGame(g.ID)
	})
}

func (gm *GameManager) cleanupGame(gameID string) {
	g, exists := gm.games[gameID]
	if !exists {
		return
	}

	delete(gm.games, gameID)
	delete(gm.playerGames, g.Player1.Username)
	if !g.IsBot {
		delete(gm.playerGames, g.Player2.Username)
	}

	kafka.SendAnalytics("game_cleaned_up", map[string]interface{}{"gameId": gameID})
	log.Printf("Game %s cleaned up", gameID)
}

func (gm *GameManager) HandleDisconnect(username string) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	// Remove from matchmaking queue
	for i, p := range gm.matchmakingQueue {
		if p.Username == username {
			gm.matchmakingQueue = append(gm.matchmakingQueue[:i], gm.matchmakingQueue[i+1:]...)
			log.Printf("Player %s removed from matchmaking queue", username)
			return
		}
	}

	// Handle in-game disconnection
	gameID, exists := gm.playerGames[username]
	if !exists {
		return
	}

	g, found := gm.games[gameID]
	if !found {
		return
	}

	g.DisconnectedPlayers[username] = true
	log.Printf("Player %s disconnected from game %s", username, gameID)

	// Notify opponent
	gm.broadcastToGame(g, map[string]interface{}{
		"type":    "playerDisconnected",
		"message": fmt.Sprintf("%s disconnected. They have 30 seconds to reconnect.", username),
	})

	// Set reconnection timer
	timer := time.AfterFunc(30*time.Second, func() {
		gm.mu.Lock()
		defer gm.mu.Unlock()

		if g.DisconnectedPlayers[username] {
			// Player didn't reconnect, end game
			g.Status = "completed"
			opponentUsername := g.Player1.Username
			if username == g.Player1.Username {
				opponentUsername = g.Player2.Username
			}
			g.Winner = opponentUsername

			gm.broadcastToGame(g, map[string]interface{}{
				"type":    "gameOver",
				"message": fmt.Sprintf("%s didn't reconnect. %s wins!", username, opponentUsername),
				"winner":  g.Winner,
			})

			gm.cleanupGame(gameID)
		}
	})

	gm.reconnectionTimers[username] = timer
}
