package main

import (
	"connect-four/database"
	"connect-four/kafka"
	"connect-four/manager"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins
		},
	}
	gameManager *manager.GameManager
	db          *database.DB
)

type WSMessage struct {
	Type     string `json:"type"`
	Username string `json:"username"`
	Col      int    `json:"col"`
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	log.Println("New WebSocket connection")

	var currentUsername string

	for {
		var msg WSMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			if currentUsername != "" {
				gameManager.HandleDisconnect(currentUsername)
			}
			break
		}

		switch msg.Type {
		case "joinMatchmaking":
			currentUsername = msg.Username
			result := gameManager.AddToMatchmaking(msg.Username, conn)

			if !result.Success {
				conn.WriteJSON(map[string]interface{}{
					"type":  "error",
					"error": result.Error,
				})
			} else if result.Type == "waiting" {
				conn.WriteJSON(map[string]interface{}{
					"type":    "matchmaking",
					"message": "Waiting for opponent...",
				})
			}

		case "makeMove":
			result := gameManager.MakeMove(msg.Username, msg.Col)

			if !result.Success {
				conn.WriteJSON(map[string]interface{}{
					"type":  "error",
					"error": result.Error,
				})
			}

		default:
			conn.WriteJSON(map[string]interface{}{
				"type":  "error",
				"error": "Unknown message type",
			})
		}
	}
}

func getLeaderboard(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 10
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	leaderboard, err := db.GetLeaderboard(limit, offset)
	if err != nil {
		log.Printf("Error fetching leaderboard: %v", err)
		leaderboard = []database.LeaderboardEntry{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"leaderboard": leaderboard,
		"limit":       limit,
		"offset":      offset,
	})
}

func getPlayerStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]

	stats, err := db.GetPlayerStats(username)
	if err != nil {
		log.Printf("Error fetching player stats: %v", err)
		stats = &database.LeaderboardEntry{
			Username:   username,
			Wins:       0,
			Losses:     0,
			Draws:      0,
			TotalGames: 0,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"stats":   stats,
	})
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"message": "Connect Four server is running",
	})
}

func main() {
	// Load environment variables
	godotenv.Load()

	// Initialize Kafka
	if err := kafka.InitKafka(); err != nil {
		log.Printf("Kafka initialization warning: %v", err)
		log.Println("Analytics will be disabled")
	}
	defer kafka.Close()

	// Initialize database
	var err error
	db, err = database.NewDB()
	if err != nil {
		log.Printf("Database initialization warning: %v", err)
		log.Println("Game will continue without database persistence")
	}

	// Initialize game manager
	gameManager = manager.NewGameManager(db)

	// Setup router
	r := mux.NewRouter()

	// WebSocket endpoint
	r.HandleFunc("/ws", handleWebSocket)

	// REST API endpoints
	r.HandleFunc("/health", healthCheck).Methods("GET")
	r.HandleFunc("/api/leaderboard", getLeaderboard).Methods("GET")
	r.HandleFunc("/api/leaderboard/player/{username}", getPlayerStats).Methods("GET")

	// CORS
	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}).Handler(r)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	log.Printf("Server starting on port %s", port)
	log.Printf("WebSocket endpoint: ws://localhost:%s/ws", port)
	log.Printf("Health check: http://localhost:%s/health", port)
	log.Printf("Leaderboard API: http://localhost:%s/api/leaderboard", port)

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
