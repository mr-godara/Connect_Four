# System Architecture - Connect Four Multiplayer

## 📐 Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLIENT LAYER                             │
├─────────────────────────────────────────────────────────────────┤
│  Browser (React Frontend)                                        │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  - Username Entry                                         │  │
│  │  - 7×6 Game Board                                         │  │
│  │  - Real-time Move Display                                 │  │
│  │  - Leaderboard                                            │  │
│  │  - Game Status                                            │  │
│  └──────────────────────────────────────────────────────────┘  │
│                          ↕ WebSocket                             │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                      APPLICATION LAYER                           │
├─────────────────────────────────────────────────────────────────┤
│  Node.js Backend Server (Express + WebSocket)                   │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  WebSocket Handler                                        │  │
│  │  ├─ Connection Management                                 │  │
│  │  ├─ Message Routing                                       │  │
│  │  └─ Event Broadcasting                                    │  │
│  └──────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  GameManager Service                                      │  │
│  │  ├─ Matchmaking Queue                                     │  │
│  │  ├─ Active Games Map                                      │  │
│  │  ├─ Player-Game Mapping                                   │  │
│  │  ├─ Reconnection Timers                                   │  │
│  │  └─ Game Lifecycle Management                             │  │
│  └──────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Game Logic                                               │  │
│  │  ├─ Board: 7×6 grid, move validation, win detection      │  │
│  │  ├─ Bot: Strategic AI with priorities                     │  │
│  │  └─ Game: State management, turn handling                 │  │
│  └──────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  REST API Routes                                          │  │
│  │  ├─ GET /health                                           │  │
│  │  ├─ GET /api/leaderboard                                  │  │
│  │  └─ GET /api/leaderboard/player/:username                 │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                        DATA LAYER                                │
├─────────────────────────────────────────────────────────────────┤
│  ┌──────────────────┐  ┌──────────────────┐                    │
│  │  PostgreSQL      │  │  Kafka           │                    │
│  │  ┌────────────┐  │  │  ┌────────────┐  │                    │
│  │  │ games      │  │  │  │ Analytics  │  │                    │
│  │  │ table      │  │  │  │ Events     │  │                    │
│  │  ├────────────┤  │  │  └────────────┘  │                    │
│  │  │leaderboard │  │  │    (Optional)    │                    │
│  │  │ table      │  │  │                  │                    │
│  │  └────────────┘  │  └──────────────────┘                    │
│  └──────────────────┘                                           │
└─────────────────────────────────────────────────────────────────┘
```

## 🔄 Data Flow Diagrams

### Player vs Player Game Flow

```
Player 1                Backend                 Player 2
   │                      │                         │
   ├──joinMatchmaking────>│                         │
   │                      ├─[Queue: Player1]        │
   │                      │                         │
   │                      │<───joinMatchmaking──────┤
   │                      ├─[Match Found!]          │
   │                      ├─[Create Game]           │
   │<───gameStart─────────┤────gameStart───────────>│
   │                      │                         │
   ├──makeMove(col=3)────>│                         │
   │                      ├─[Validate Turn]         │
   │                      ├─[Apply Move]            │
   │                      ├─[Check Winner]          │
   │<───move──────────────┤────move────────────────>│
   │                      │                         │
   │                      │<───makeMove(col=4)──────┤
   │<───move──────────────┤────move────────────────>│
   │                      │                         │
   │      [Game continues until winner or draw]     │
   │                      │                         │
   │<───gameOver──────────┤────gameOver────────────>│
   │                      ├─[Save to DB]            │
   │                      ├─[Update Leaderboard]    │
   │                      └─[Emit Kafka Event]      │
```

### Player vs Bot Game Flow

```
Player                  Backend                  Bot AI
   │                      │                        │
   ├──joinMatchmaking────>│                        │
   │                      ├─[Queue: Player]        │
   │                      ├─[Start 10s Timer]      │
   │                      │                        │
   │      [10 seconds pass - no 2nd player]        │
   │                      │                        │
   │                      ├─[Timeout!]             │
   │                      ├─[Create Bot Game]      │
   │<───gameStart─────────┤                        │
   │                      │                        │
   ├──makeMove(col=3)────>│                        │
   │                      ├─[Apply Move]           │
   │<───move──────────────┤                        │
   │                      │                        │
   │                      ├─[Bot's Turn]           │
   │                      ├───getBestMove()───────>│
   │                      │<──return col=3─────────┤
   │                      ├─[Apply Bot Move]       │
   │<───move──────────────┤                        │
   │                      │                        │
```

### Reconnection Flow

```
Player              Backend              Database
   │                  │                     │
   │─[Connected]──────│                     │
   │                  │                     │
   │ [Disconnect!]    │                     │
   │                  ├─[Mark Disconnected] │
   │                  ├─[Start 30s Timer]   │
   │                  │                     │
   │─[Reconnect]─────>│                     │
   │                  ├─[Clear Timer]       │
   │                  ├─[Restore State]     │
   │<──gameState──────┤                     │
   │                  │                     │
```

## 🎮 Game State Machine

```
┌─────────────┐
│   INIT      │
└──────┬──────┘
       │
       v
┌─────────────┐    No Match After 10s
│ MATCHMAKING ├───────────────────┐
└──────┬──────┘                   │
       │ Match Found              │
       v                          v
┌─────────────┐            ┌─────────────┐
│  PVP GAME   │            │  BOT GAME   │
│   ACTIVE    │            │   ACTIVE    │
└──────┬──────┘            └──────┬──────┘
       │                          │
       │ Player 1 Move            │ Player Move
       v                          v
┌─────────────┐            ┌─────────────┐
│   PLAYER    │            │   PLAYER    │
│  2's TURN   │            │   TURN      │
└──────┬──────┘            └──────┬──────┘
       │                          │
       │                          v
       │                   ┌─────────────┐
       │                   │   BOT       │
       │                   │   TURN      │
       │                   └──────┬──────┘
       │                          │
       v                          v
┌─────────────────────────────────────┐
│         CHECK WIN/DRAW              │
└──────┬────────────────────┬─────────┘
       │ Continue           │ Game Over
       v                    v
   [Loop Back]      ┌─────────────┐
                    │  COMPLETED  │
                    │  ├─Save DB  │
                    │  ├─Update   │
                    │  └─Cleanup  │
                    └─────────────┘
```

## 🔌 WebSocket Message Types

### Client → Server

```javascript
// Join matchmaking
{
  type: "joinMatchmaking",
  username: "Player1"
}

// Make a move
{
  type: "makeMove",
  username: "Player1",
  col: 3  // 0-6
}

// Reconnect to game
{
  type: "reconnect",
  username: "Player1"
}
```

### Server → Client

```javascript
// Matchmaking status
{
  type: "matchmaking",
  message: "Waiting for opponent..."
}

// Game started
{
  type: "gameStart",
  game: {
    id: "uuid",
    player1: "Player1",
    player2: "Player2",
    board: [[0,0,...], ...],
    currentTurn: 1,
    isBot: false
  },
  playerNumber: 1
}

// Move made
{
  type: "move",
  username: "Player1",
  playerNumber: 1,
  col: 3,
  row: 5,
  board: [[...], ...],
  nextTurn: 2,
  gameOver: false
}

// Game over
{
  type: "gameOver",
  winner: "Player1", // or "draw"
  reason: "win", // or "forfeit"
  message: "Player1 wins!"
}

// Player disconnected
{
  type: "playerDisconnected",
  username: "Player2",
  message: "Player2 disconnected..."
}

// Player reconnected
{
  type: "playerReconnected",
  username: "Player2",
  message: "Player2 reconnected!"
}

// Reconnection successful
{
  type: "reconnected",
  game: {...},
  playerNumber: 1
}

// Error
{
  type: "error",
  error: "Not your turn"
}
```

## 💾 Database Schema

### Games Table

```sql
CREATE TABLE games (
  id VARCHAR(255) PRIMARY KEY,
  player1 VARCHAR(255) NOT NULL,
  player2 VARCHAR(255) NOT NULL,
  winner VARCHAR(255),              -- player name or "draw"
  game_duration INTEGER,            -- seconds
  is_bot_game BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  completed_at TIMESTAMP
);

-- Indexes
CREATE INDEX idx_games_player1 ON games(player1);
CREATE INDEX idx_games_player2 ON games(player2);
CREATE INDEX idx_games_created_at ON games(created_at DESC);
```

### Leaderboard Table

```sql
CREATE TABLE leaderboard (
  username VARCHAR(255) PRIMARY KEY,
  wins INTEGER DEFAULT 0,
  losses INTEGER DEFAULT 0,
  draws INTEGER DEFAULT 0,
  total_games INTEGER DEFAULT 0,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_leaderboard_wins ON leaderboard(wins DESC);
CREATE INDEX idx_leaderboard_total_games ON leaderboard(total_games DESC);
```

## 📊 Kafka Events Schema

```javascript
{
  eventType: "player_joined_matchmaking" | "game_created" | 
             "bot_game_started" | "move_made" | "bot_move" |
             "game_completed" | "player_disconnected" | 
             "player_reconnected" | "game_forfeited" | 
             "game_cleaned_up",
  timestamp: "2026-01-29T12:00:00.000Z",
  data: {
    // Event-specific data
    username?: string,
    gameId?: string,
    player1?: string,
    player2?: string,
    col?: number,
    row?: number,
    winner?: string,
    duration?: number,
    isBot?: boolean
  }
}
```

## 🧩 Component Architecture

### Backend Services

```
GameManager (Central Service)
├── Matchmaking Queue Management
│   ├── Add player to queue
│   ├── Match two players
│   └── Timeout handling
├── Game Management
│   ├── Create game
│   ├── Process moves
│   ├── Handle bot moves
│   └── End game
├── Player Management
│   ├── Track player-game mapping
│   ├── Handle disconnections
│   ├── Reconnection logic
│   └── Forfeit handling
└── Data Persistence
    ├── Save games to DB
    └── Update leaderboard

Game (Game Instance)
├── Board Management
├── Turn Management
├── Win Detection
├── Bot Integration
└── State Export

Board (Game Logic)
├── Move Validation
├── Disc Placement
├── Win Detection (4 directions)
├── Draw Detection
└── State Management

Bot (AI Logic)
├── Win Move Detection
├── Block Opponent
├── Center Preference
└── Safe Move Selection
```

### Frontend Components

```
App (Main Component)
├── Login Screen
│   └── Username Input
├── Game Screen
│   ├── Status Bar
│   │   ├── Player Info
│   │   ├── Opponent Info
│   │   └── Turn Indicator
│   ├── Game Container
│   │   ├── Board Section
│   │   │   ├── Column Buttons
│   │   │   └── Board Grid
│   │   │       └── Cell Components
│   │   └── Sidebar
│   │       ├── Game Info
│   │       └── Leaderboard
│   └── Game Over Modal
│       ├── Winner Display
│       └── Action Buttons
└── Message Banner
```

## 🔐 Security Considerations

### Implemented
- ✅ Parameterized SQL queries (injection prevention)
- ✅ Input validation (username, moves)
- ✅ CORS configuration
- ✅ Error message sanitization
- ✅ Turn validation (prevent cheating)

### Not Implemented (Future)
- ⚠️ WebSocket authentication
- ⚠️ Rate limiting
- ⚠️ DDoS protection
- ⚠️ HTTPS/WSS in production

## 📈 Scalability Considerations

### Current Design
- Single server instance
- In-memory game state
- Works for: 100s of concurrent games

### Scaling Strategy
1. **Horizontal Scaling**
   - Use Redis for shared state
   - Sticky sessions on load balancer
   - Or use Redis pub/sub for game events

2. **Database Scaling**
   - Connection pooling (already implemented)
   - Read replicas for leaderboard
   - Sharding by game ID

3. **Kafka Scaling**
   - Already decoupled
   - Multiple brokers
   - Partitioning by game ID

## 🔧 Technology Stack Details

```
Frontend:
├── React 18.2.0
├── WebSocket API (native)
└── CSS3 (custom styling)

Backend:
├── Node.js 18+
├── Express 4.18.2
├── ws 8.14.2 (WebSocket)
├── pg 8.11.3 (PostgreSQL client)
├── kafkajs 2.2.4
├── uuid 9.0.1
├── cors 2.8.5
└── dotenv 16.3.1

Database:
├── PostgreSQL 15
└── Kafka 2.x + Zookeeper

DevOps:
├── Docker
├── Docker Compose
└── Node Alpine images
```

## 📦 Deployment Topology

### Development
```
Developer Machine
├── Backend (localhost:3001)
├── Frontend (localhost:3000)
├── PostgreSQL (localhost:5432)
└── Kafka (localhost:9092)
```

### Docker Deployment
```
Docker Network
├── postgres container
│   └── Port: 5432
├── zookeeper container
│   └── Port: 2181
├── kafka container
│   └── Port: 9092
├── backend container
│   └── Port: 3001
└── frontend container
    └── Port: 3000
```

### Production (Suggested)
```
Load Balancer
├── Backend Cluster
│   ├── Backend Instance 1
│   ├── Backend Instance 2
│   └── Backend Instance N
├── Static CDN
│   └── Frontend Build
├── Database Cluster
│   ├── Primary PostgreSQL
│   └── Read Replicas
├── Kafka Cluster
│   ├── Broker 1
│   ├── Broker 2
│   └── Broker 3
└── Redis Cluster
    └── Shared State Management
```

This architecture provides a robust, scalable, and maintainable solution for real-time multiplayer Connect Four!
