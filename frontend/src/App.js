import React, { useState, useEffect, useCallback } from 'react';
import './index.css';

const WS_URL = 'wss://connect-four-backend-386a.onrender.com/ws';
const API_URL = 'https://connect-four-backend-386a.onrender.com';

function App() {
  const [username, setUsername] = useState('');
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [ws, setWs] = useState(null);
  const [gameState, setGameState] = useState(null);
  const [playerNumber, setPlayerNumber] = useState(null);
  const [message, setMessage] = useState({ text: '', type: '' });
  const [leaderboard, setLeaderboard] = useState([]);
  const [gameOver, setGameOver] = useState(false);
  const [winner, setWinner] = useState(null);

  // Fetch leaderboard
  const fetchLeaderboard = useCallback(async () => {
    try {
      const response = await fetch(`${API_URL}/leaderboard?limit=10`);
      const data = await response.json();
      if (data.success) {
        setLeaderboard(data.leaderboard);
      }
    } catch (error) {
      console.error('Error fetching leaderboard:', error);
    }
  }, []);

  // Connect to WebSocket
  useEffect(() => {
    if (!isLoggedIn) return;

    const websocket = new WebSocket(WS_URL);

    websocket.onopen = () => {
      console.log('Connected to WebSocket');
      setMessage({ text: 'Connected! Joining matchmaking...', type: 'success' });
      
      // Auto-join matchmaking once connected
      websocket.send(JSON.stringify({
        type: 'joinMatchmaking',
        username,
      }));
    };

    websocket.onmessage = (event) => {
      const data = JSON.parse(event.data);
      handleWebSocketMessage(data);
    };

    websocket.onerror = (error) => {
      console.error('WebSocket error:', error);
      setMessage({ text: 'Connection error', type: 'error' });
    };

    websocket.onclose = () => {
      console.log('WebSocket connection closed');
      setMessage({ text: 'Disconnected from server', type: 'warning' });
    };

    setWs(websocket);

    return () => {
      if (websocket.readyState === WebSocket.OPEN) {
        websocket.close();
      }
    };
  }, [isLoggedIn, username]);

  // Handle WebSocket messages
  const handleWebSocketMessage = (data) => {
    console.log('Received:', data);

    switch (data.type) {
      case 'matchmaking':
        setMessage({ text: data.message, type: 'info' });
        break;

      case 'gameStart':
        setGameState(data.game);
        setPlayerNumber(data.playerNumber);
        setGameOver(false);
        setWinner(null);
        const opponentName = data.game.isBot ? 'Bot' : 
          (data.playerNumber === 1 ? data.game.player2 : data.game.player1);
        setMessage({ 
          text: data.message || `Game started! You are Player ${data.playerNumber} (${data.playerNumber === 1 ? 'Red' : 'Yellow'}). Playing against: ${opponentName}`, 
          type: 'success' 
        });
        break;

      case 'move':
        setGameState(prev => ({
          ...prev,
          board: data.board,
          currentTurn: data.nextTurn,
        }));

        if (data.gameOver) {
          setGameOver(true);
          setWinner(data.winner);
          fetchLeaderboard();
          
          // Display game over message
          let gameOverMsg = '';
          if (data.winner === 'draw') {
            gameOverMsg = "🤝 Game Over - It's a Draw!";
          } else if (data.winner === username) {
            gameOverMsg = `🎉 You Win! ${data.username} connected 4 in column ${data.col + 1}!`;
          } else {
            gameOverMsg = `😔 You Lost! ${data.winner} connected 4 in column ${data.col + 1}!`;
          }
          setMessage({ text: gameOverMsg, type: data.winner === username ? 'success' : 'warning' });
        } else {
          const turnText = data.nextTurn === playerNumber ? 'Your turn!' : `Opponent's turn`;
          setMessage({ text: `${data.username} played column ${data.col + 1}. ${turnText}`, type: 'info' });
        }
        break;

      case 'reconnected':
        setGameState(data.game);
        setPlayerNumber(data.playerNumber);
        setMessage({ text: 'Reconnected successfully!', type: 'success' });
        break;

      case 'playerDisconnected':
        setMessage({ text: data.message, type: 'warning' });
        break;

      case 'playerReconnected':
        setMessage({ text: data.message, type: 'success' });
        break;

      case 'gameOver':
        setGameOver(true);
        setWinner(data.winner);
        setMessage({ text: data.message || `Game over! Winner: ${data.winner}`, type: 'info' });
        fetchLeaderboard();
        break;

      case 'error':
        setMessage({ text: data.error, type: 'error' });
        break;

      default:
        console.log('Unknown message type:', data.type);
    }
  };

  // Join matchmaking
  const joinMatchmaking = () => {
    if (!username.trim()) {
      setMessage({ text: 'Please enter a username', type: 'error' });
      return;
    }

    setIsLoggedIn(true);
    fetchLeaderboard();
  };

  // Make a move
  const makeMove = (col) => {
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      setMessage({ text: 'Not connected to server', type: 'error' });
      return;
    }

    if (gameState.currentTurn !== playerNumber) {
      setMessage({ text: 'Not your turn!', type: 'error' });
      return;
    }

    ws.send(JSON.stringify({
      type: 'makeMove',
      username,
      col,
    }));
  };

  // Play again
  const playAgain = () => {
    // Clear current game state
    setGameState(null);
    setPlayerNumber(null);
    setGameOver(false);
    setWinner(null);
    setMessage({ text: 'Searching for a new game...', type: 'info' });
    
    // Wait a moment to ensure state is cleared, then join matchmaking
    setTimeout(() => {
      if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({
          type: 'joinMatchmaking',
          username,
        }));
      }
    }, 100);
  };

  // Render login screen
  if (!isLoggedIn) {
    return (
      <div className="app">
        <div className="container">
          <h1>🔴 Connect Four 🟡</h1>
          <div className="login-screen">
            <input
              type="text"
              placeholder="Enter your username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              onKeyPress={(e) => e.key === 'Enter' && joinMatchmaking()}
            />
            <button onClick={joinMatchmaking}>Join Game</button>
          </div>
        </div>
      </div>
    );
  }

  // Render game screen
  return (
    <div className="app">
      <div className="container">
        <h1>🔴 Connect Four 🟡</h1>

        {message.text && (
          <div className={`message ${message.type}`}>
            {message.text}
          </div>
        )}

        {gameState && (
          <>
            <div className="status-bar">
              <p><strong>You:</strong> {username} (Player {playerNumber}) - {playerNumber === 1 ? '🔴' : '🟡'}</p>
              <p><strong>Opponent:</strong> {playerNumber === 1 ? gameState.player2 : gameState.player1}</p>
              {!gameOver && (
                <p className="current-turn">
                  {gameState.currentTurn === playerNumber ? '🎯 YOUR TURN' : '⏳ Opponent\'s Turn'}
                </p>
              )}
            </div>

            <div className="game-container">
              <div className="board-section">
                <div className="column-buttons">
                  {[0, 1, 2, 3, 4, 5, 6].map((col) => (
                    <button
                      key={col}
                      className="column-button"
                      onClick={() => makeMove(col)}
                      disabled={gameOver || gameState.currentTurn !== playerNumber}
                    >
                      {col + 1}
                    </button>
                  ))}
                </div>

                <div className="board">
                  {gameState.board.map((row, rowIndex) => (
                    <div key={rowIndex} className="board-row">
                      {row.map((cell, colIndex) => (
                        <div
                          key={colIndex}
                          className={`cell ${gameOver || gameState.currentTurn !== playerNumber ? 'disabled' : ''}`}
                          onClick={() => !gameOver && gameState.currentTurn === playerNumber && makeMove(colIndex)}
                        >
                          {cell !== 0 && (
                            <div className={`disc player${cell}`}></div>
                          )}
                        </div>
                      ))}
                    </div>
                  ))}
                </div>
              </div>

              <div className="sidebar">
                <div className="game-info">
                  <h2>Game Info</h2>
                  <p><strong>Game ID:</strong> {gameState.id.substring(0, 8)}...</p>
                  <p><strong>Status:</strong> {gameOver ? 'Completed' : 'Active'}</p>
                  <p><strong>Mode:</strong> {gameState.isBot ? 'vs Bot' : 'vs Player'}</p>
                </div>

                <div className="leaderboard">
                  <h2>🏆 Leaderboard</h2>
                  {leaderboard.length > 0 ? (
                    <table className="leaderboard-table">
                      <thead>
                        <tr>
                          <th>#</th>
                          <th>Player</th>
                          <th>Wins</th>
                        </tr>
                      </thead>
                      <tbody>
                        {leaderboard.map((player, index) => (
                          <tr key={player.username} style={{
                            backgroundColor: player.username === username ? '#e3f2fd' : 'transparent'
                          }}>
                            <td>{index + 1}</td>
                            <td>{player.username}</td>
                            <td>{player.wins}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  ) : (
                    <p>No data yet</p>
                  )}
                </div>
              </div>
            </div>
          </>
        )}

        {gameOver && (
          <div className="game-over-modal">
            <div className="modal-content">
              <h2>Game Over!</h2>
              <p style={{ fontSize: '48px' }}>
                {winner === 'draw' ? '🤝' : 
                 winner === username ? '🎉' : '😢'}
              </p>
              <p>
                {winner === 'draw' ? "It's a draw!" :
                 winner === username ? 'You won!' : 
                 `${winner} won!`}
              </p>
              <button onClick={playAgain}>Play Again</button>
              <button onClick={() => {
                setIsLoggedIn(false);
                setGameState(null);
                setPlayerNumber(null);
                setGameOver(false);
                setWinner(null);
                if (ws) ws.close();
              }}>
                Exit
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

export default App;
