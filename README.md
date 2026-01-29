# Connect Four - Real-Time Multiplayer Game

A real-time multiplayer Connect Four game with **Go backend**, WebSocket support, competitive bot AI, player matchmaking, reconnection handling, and PostgreSQL persistence.

## 🎮 Features

- **Real-time multiplayer gameplay** using WebSockets
- **Automatic matchmaking** with 10-second timeout
- **Competitive bot AI** (non-random, strategic play)
- **Player reconnection** support (30-second grace period)
- **Leaderboard system** with persistent stats
- **PostgreSQL** for game history and leaderboard
- **React frontend** with clean, functional UI
- **Go backend** - Fast, compiled, production-ready

## 📁 Project Structure

```
Emitrr_Assignment/
├── backend/                # Go backend server
│   ├── game/              # Game logic (Board, Bot, Game)
│   ├── manager/           # GameManager service
│   ├── database/          # PostgreSQL integration
│   ├── main.go            # Main server file
│   └── go.mod             # Go dependencies
│
├── frontend/              # React frontend
│   ├── public/
│   ├── src/
│   │   ├── App.js        # Main React component
│   │   ├── index.js
│   │   └── index.css
│   └── package.json
```

## 🚀 Quick Start

### Prerequisites

- **Go 1.21+** (Download: https://go.dev/dl/)
- **PostgreSQL 15+** (optional - game works without it)
- **Node.js 18+** (for frontend only)
- npm or yarn

### Setup Instructions

#### Backend Setup (Go)

1. **Navigate to backend:**
   ```bash
   cd backend
   ```

2. **Download dependencies:**
   ```bash
   go mod download
   ```

3. **Set up PostgreSQL database (optional):**
   ```sql
   CREATE DATABASE connect_four;
   ```

4. **Configure environment variables:**
   
   Copy `.env.example` to `.env` and update:
   ```
   PORT=3001
   DATABASE_URL=postgresql://postgres:password@localhost:5432/connect_four?sslmode=disable
   
   # For Render PostgreSQL:
   # DATABASE_URL=postgresql://user:password@host.render.com:5432/dbname
   
   MATCHMAKING_TIMEOUT=10000
   RECONNECTION_TIMEOUT=30000
   ```

5. **Start backend server:**
   ```bash
   go run main.go
   
   # Or build binary:
   go build -o server main.go
   ./server
   ```

#### Frontend Setup

1. **Navigate to frontend:**
   ```bash
   cd frontend
   ```

2. **Install dependencies:**
   ```bash
   npm install
   ```

3. **Start frontend:**
   ```bash
   npm start
   ```

4. **Access the game:**
   - Open http://localhost:3000 in your browser

## 🎯 How to Play

1. **Enter your username** on the login screen
2. **Click "Join Game"** to enter matchmaking
3. **Wait for opponent** (max 10 seconds) or play against bot
4. **Drop discs** by clicking column buttons or cells
5. **Win by connecting 4** discs horizontally, vertically, or diagonally
6. **View leaderboard** to see top players

## 🧠 Game Rules

- Board: 7 columns × 6 rows
- Players take turns dropping discs
- Discs fall to the lowest empty position
- First to connect 4 discs wins
- Full board with no winner = Draw

## 🤖 Bot AI Strategy

The bot uses rule-based logic with priorities:

1. **Win immediately** if possible
2. **Block opponent's** winning move
3. **Prefer center columns** (3, 2, 4, 1, 5, 0, 6)
4. **Avoid creating** opponent winning opportunities

## 🔌 API Endpoints

### REST API

- `GET /health` - Health check
- `GET /api/leaderboard` - Get top players
- `GET /api/leaderboard/player/:username` - Get player stats

### WebSocket Messages

**Client → Server:**
- `joinMatchmaking` - Join matchmaking queue
- `makeMove` - Make a move (col)
- `reconnect` - Reconnect to game

**Server → Client:**
- `matchmaking` - Matchmaking status
- `gameStart` - Game started
- `move` - Move made
- `gameOver` - Game ended
- `playerDisconnected` - Player disconnected
- `playerReconnected` - Player reconnected
- `error` - Error message

## 📊 Database Schema

### Games Table
```sql
id VARCHAR(255) PRIMARY KEY
player1 VARCHAR(255)
player2 VARCHAR(255)
winner VARCHAR(255)
game_duration INTEGER
is_bot_game BOOLEAN
created_at TIMESTAMP
completed_at TIMESTAMP
```

### Leaderboard Table
```sql
username VARCHAR(255) PRIMARY KEY
wins INTEGER
losses INTEGER
draws INTEGER
total_games INTEGER
updated_at TIMESTAMP
```

## 📈 Kafka Analytics Events

The system emits the following events:

- `player_joined_matchmaking`
- `game_created`
- `bot_game_started`
- `move_made`
- `bot_move`
- `game_completed`
- `player_disconnected`
- `player_reconnected`
- `game_forfeited`

**Note:** Kafka is optional. If unavailable, the system logs events and continues without analytics.

## 🔄 Reconnection Flow

1. Player disconnects (network issue, browser close, etc.)
2. System waits 30 seconds for reconnection
3. Player can reconnect using same username
4. Game state is restored
5. If timeout expires, opponent wins by forfeit

## 🛠️ Tech Stack

**Backend:**
- Go 1.21+
- Gorilla WebSocket
- Gorilla Mux (HTTP router)
- PostgreSQL (lib/pq driver)
- CORS middleware

**Frontend:**
- React 18
- WebSocket API
- CSS3

**Database:**
- PostgreSQL 15+

## ⚡ Why Go?

- **10x faster** than Node.js - Compiled binary with native concurrency
- **Type-safe** - Compile-time error detection
- **Single binary** - No dependencies, easy deployment
- **Better concurrency** - Goroutines handle thousands of WebSocket connections
- **Lower memory** - ~10MB vs Node.js ~50MB
- **Production-ready** - Used by Docker, Kubernetes, YouTube

## 🧪 Testing

### Test Matchmaking

1. Open two browser windows
2. Enter different usernames
3. Join matchmaking in both
4. Game should start within seconds

### Test Bot Game

1. Open one browser window
2. Join matchmaking
3. Wait 10 seconds
4. Bot game should start

### Test Reconnection

1. Start a game
2. Close the browser tab
3. Reopen and enter same username
4. Should reconnect to active game


## 📄 License

MIT

## 👤 Author

Built for Emitrr Assignment

---

**Enjoy the game! 🎮**
