# Quiz Game Test Script
Write-Host "`n========== Quiz Game Test Script ==========" -ForegroundColor Cyan

# Check if server is running
$serverProcess = Get-Process -Name "server" -ErrorAction SilentlyContinue
if ($serverProcess) {
    Write-Host "✓ Server is running (PID: $($serverProcess.Id))" -ForegroundColor Green
} else {
    Write-Host "✗ Server is not running. Starting server..." -ForegroundColor Yellow
    Start-Process -FilePath ".\bin\server.exe" -WorkingDirectory "." -NoNewWindow
    Start-Sleep -Seconds 3
    Write-Host "✓ Server started" -ForegroundColor Green
}

# Test server health
Write-Host "`nTesting server health..." -ForegroundColor Cyan
try {
    $health = Invoke-RestMethod -Uri "http://localhost:8080/health" -Method Get
    Write-Host "✓ Server is healthy: $($health.status)" -ForegroundColor Green
} catch {
    Write-Host "✗ Server health check failed: $_" -ForegroundColor Red
    exit 1
}

# Create game room
Write-Host "`nCreating QA game room..." -ForegroundColor Cyan
$createRoomQuery = @"
mutation {
  createGameRoom(input: {
    name: "Test QA Room"
    gameType: QA
    hostUsername: "testuser"
    maxPlayers: 4
    totalRounds: 3
    roundTimeLimit: 15
  }) {
    id
    name
    gameType
    status
  }
}
"@

try {
    $createResponse = Invoke-RestMethod -Uri "http://localhost:8080/graphql" -Method Post -Body (@{query=$createRoomQuery} | ConvertTo-Json) -ContentType "application/json"
    $roomId = $createResponse.data.createGameRoom.id
    Write-Host "✓ Room created: $roomId" -ForegroundColor Green
    Write-Host "  Name: $($createResponse.data.createGameRoom.name)"
    Write-Host "  Type: $($createResponse.data.createGameRoom.gameType)"
    Write-Host "  Status: $($createResponse.data.createGameRoom.status)"
} catch {
    Write-Host "✗ Failed to create room: $_" -ForegroundColor Red
    exit 1
}

# Join room with second player
Write-Host "`nJoining room with second player..." -ForegroundColor Cyan
$joinRoomQuery = @"
mutation {
  joinGameRoom(roomId: "$roomId", username: "testuser2") {
    id
    players {
      username
    }
  }
}
"@

try {
    $joinResponse = Invoke-RestMethod -Uri "http://localhost:8080/graphql" -Method Post -Body (@{query=$joinRoomQuery} | ConvertTo-Json) -ContentType "application/json"
    Write-Host "✓ Player joined. Players count: $($joinResponse.data.joinGameRoom.players.Count)" -ForegroundColor Green
} catch {
    Write-Host "✗ Failed to join room: $_" -ForegroundColor Red
}

# Start game
Write-Host "`nStarting game..." -ForegroundColor Cyan
$startGameQuery = @"
mutation {
  startGame(roomId: "$roomId") {
    id
    status
    currentRound
  }
}
"@

try {
    $startResponse = Invoke-RestMethod -Uri "http://localhost:8080/graphql" -Method Post -Body (@{query=$startGameQuery} | ConvertTo-Json) -ContentType "application/json"
    Write-Host "✓ Game started!" -ForegroundColor Green
    Write-Host "  Status: $($startResponse.data.startGame.status)"
    Write-Host "  Round: $($startResponse.data.startGame.currentRound)"
} catch {
    Write-Host "✗ Failed to start game: $_" -ForegroundColor Red
    Write-Host "  Error: $($_.Exception.Message)"
    exit 1
}

Write-Host "`n========== Waiting for quiz to be sent ==========" -ForegroundColor Cyan
Write-Host "Check server logs for:" -ForegroundColor Yellow
Write-Host "  - [StartGame] Room $roomId game type: QA"
Write-Host "  - [Quiz] ==================== START QUIZ GAME ===================="
Write-Host "  - [Quiz] ########## SEND NEXT QUIZ ##########"
Write-Host "  - [Quiz] SUCCESS: Got QA quiz ID=xxx"

Start-Sleep -Seconds 5

Write-Host "`n========== Test Complete ==========" -ForegroundColor Cyan
