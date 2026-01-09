# OX Quiz Test Script
Write-Host "========== OX Quiz Game Test Script ==========" -ForegroundColor Cyan

# Check if server is running
$serverProcess = Get-Process -Name "server" -ErrorAction SilentlyContinue
if (-not $serverProcess) {
    Write-Host "❌ Server is not running!" -ForegroundColor Red
    exit 1
}
Write-Host "✅ Server is running (PID: $($serverProcess.Id))" -ForegroundColor Green

# Test server health
Write-Host "`nTesting server health..." -ForegroundColor Yellow
$healthResponse = Invoke-RestMethod -Uri "http://localhost:8080/health" -Method GET
Write-Host "✅ Server is healthy: $($healthResponse.status)" -ForegroundColor Green

# GraphQL endpoint
$graphqlUrl = "http://localhost:8080/graphql"

# Create OX game room
Write-Host "`nCreating OX game room..." -ForegroundColor Yellow
$createRoomQuery = @{
    query = @"
mutation {
  createGameRoom(
    input: {
      name: "Test OX Room"
      gameType: OX
      maxPlayers: 4
      roundTimeLimit: 15
      totalRounds: 3
      hostUsername: "testuser"
    }
  ) {
    id
    name
    gameType
    status
  }
}
"@
} | ConvertTo-Json

$createResponse = Invoke-RestMethod -Uri $graphqlUrl -Method POST -Body $createRoomQuery -ContentType "application/json"
$roomId = $createResponse.data.createGameRoom.id
$roomName = $createResponse.data.createGameRoom.name
$gameType = $createResponse.data.createGameRoom.gameType
$status = $createResponse.data.createGameRoom.status

Write-Host "✅ Room created: $roomId" -ForegroundColor Green
Write-Host "  Name: $roomName"
Write-Host "  Type: $gameType"
Write-Host "  Status: $status"

# Join room with second player
Write-Host "`nJoining room with second player..." -ForegroundColor Yellow
$joinRoomQuery = @{
    query = @"
mutation {
  joinGameRoom(roomId: "$roomId", username: "testuser2") {
    id
    players {
      username
    }
  }
}
"@
} | ConvertTo-Json

$joinResponse = Invoke-RestMethod -Uri $graphqlUrl -Method POST -Body $joinRoomQuery -ContentType "application/json"
$playersCount = $joinResponse.data.joinGameRoom.players.Count
Write-Host "✅ Player joined. Players count: $playersCount" -ForegroundColor Green

# Start game
Write-Host "`nStarting game..." -ForegroundColor Yellow
$startGameQuery = @{
    query = @"
mutation {
  startGame(roomId: "$roomId") {
    id
    status
    currentRound
  }
}
"@
} | ConvertTo-Json

$startResponse = Invoke-RestMethod -Uri $graphqlUrl -Method POST -Body $startGameQuery -ContentType "application/json"
$gameStatus = $startResponse.data.startGame.status
$currentRound = $startResponse.data.startGame.currentRound

Write-Host "✅ Game started!" -ForegroundColor Green
Write-Host "  Status: $gameStatus"
Write-Host "  Round: $currentRound"

Write-Host "`n========== Waiting for quiz to be sent ==========" -ForegroundColor Cyan
Write-Host "Check server logs for:" -ForegroundColor Yellow
Write-Host "  - [StartGame] Room $roomId game type: OX"
Write-Host "  - [Quiz] ==================== START QUIZ GAME ===================="
Write-Host "  - [Quiz] ########## SEND NEXT QUIZ ##########"
Write-Host "  - [Quiz] SUCCESS: Got OX quiz ID=xxx"

Write-Host "`n========== Test Complete ==========" -ForegroundColor Cyan
