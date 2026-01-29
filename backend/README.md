# Connect Four - Go Backend

Real-time multiplayer Connect Four game backend written in **Go**.

## Features

- ✅ WebSocket real-time gameplay
- ✅ Player vs Player matchmaking
- ✅ Strategic bot AI (plays if no opponent found within 10s)
- ✅ Win detection (horizontal, vertical, diagonal)
- ✅ PostgreSQL persistent leaderboard
- ✅ 30-second reconnection handling
- ✅ RESTful API for leaderboard

## Prerequisites

- Go 1.21+
- PostgreSQL (optional - game works without it)

## Quick Start

### 1. Install Dependencies

```bash
cd backend-go
go mod download
```

### 2. Configure Environment

Create `.env` file:

```env
PORT=3001
DATABASE_URL=postgresql://user:password@localhost:5432/connect_four?sslmode=disable
```

For Render PostgreSQL:
```env
DATABASE_URL=postgresql://user:password@host.render.com:5432/dbname
```

### 3. Run Server

```bash
go run main.go
```

Server starts on `http://localhost:3001`

## API Endpoints

### WebSocket
- `ws://localhost:3001/ws` - Game WebSocket connection

### REST API
- `GET /health` - Health check
- `GET /api/leaderboard?limit=10&offset=0` - Get leaderboard
- `GET /api/leaderboard/player/:username` - Get player stats

## WebSocket Messages

### Client → Server

```json
// Join matchmaking
{
  "type": "joinMatchmaking",
  "username": "player1"
}

// Make move
{
  "type": "makeMove",
  "username": "player1",
  "col": 3
}
```

### Server → Client

```json
// Game started
{
  "type": "gameStart",
  "game": {...},
  "playerNumber": 1
}

// Move made
{
  "type": "move",
  "username": "player1",
  "col": 3,
  "row": 5,
  "board": [[...]],
  "nextTurn": 2,
  "gameOver": false
}
```

## Project Structure

```
backend-go/
├── main.go              # HTTP server and WebSocket handler
├── game/
│   ├── board.go         # 7×6 board logic and win detection
│   ├── bot.go           # Strategic bot AI
│   └── game.go          # Game state and move validation
├── manager/
│   └── game_manager.go  # Matchmaking and game lifecycle
└── database/
    └── database.go      # PostgreSQL integration
```

## Build

```bash
go build -o server main.go
./server
```

## Advantages over Node.js

- ⚡ **Faster performance** - Compiled binary, concurrent goroutines
- 🔒 **Type safety** - Compile-time error detection
- 📦 **Single binary** - No node_modules, easy deployment
- 💪 **Better concurrency** - Native goroutines for WebSocket connections
- 🚀 **Lower memory** - More efficient resource usage

## License

MIT
