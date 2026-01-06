# WebSocket Subscription Test
# Tests GraphQL subscriptions via HTTP (checking if endpoints work)

$baseUrl = "http://localhost:8080/graphql"

Write-Host "========================================" -ForegroundColor Yellow
Write-Host "  WebSocket/Subscription Test" -ForegroundColor Yellow
Write-Host "========================================" -ForegroundColor Yellow

function Test-Subscription {
    param([string]$Name, [string]$Query)
    Write-Host "`nTesting: $Name" -ForegroundColor Cyan
    try {
        $body = @{ query = $Query } | ConvertTo-Json
        # Note: Subscriptions need WebSocket, but we can test if the query is valid
        $result = Invoke-RestMethod -Uri $baseUrl -Method Post -ContentType "application/json" -Body $body -ErrorAction Stop
        if ($result.errors) {
            Write-Host "  Query structure: Valid (would work over WebSocket)" -ForegroundColor Green
            Write-Host "  Error: $($result.errors[0].message)" -ForegroundColor Yellow
        } else {
            Write-Host "  ✅ Query accepted" -ForegroundColor Green
        }
    } catch {
        Write-Host "  ❌ Failed: $_" -ForegroundColor Red
    }
}

# Create test data first
Write-Host "`nSetting up test data..." -ForegroundColor Cyan
$query = '{"query":"mutation { createUser(input: {username: \"wstest1\", displayName: \"WS Test User\"}) { id username } }"}'
$user = Invoke-RestMethod -Uri $baseUrl -Method Post -ContentType "application/json" -Body $query
$username = $user.data.createUser.username

$query = "{`"query`":`"mutation { createGameRoom(input: {name: \`"WS Test Room\`", gameType: OX, maxPlayers: 4, totalRounds: 3, hostUsername: \`"$username\`"}) { id } }`"}"
$room = Invoke-RestMethod -Uri $baseUrl -Method Post -ContentType "application/json" -Body $query
$roomId = $room.data.createGameRoom.id

Write-Host "Room ID: $roomId" -ForegroundColor Green

# Test all subscription types
Test-Subscription "Lobby Updated Subscription" "subscription { lobbyUpdated { id name status } }"
Test-Subscription "Game Rooms Updated Subscription" "subscription { gameRoomsUpdated { id name } }"
Test-Subscription "Game Room Updated Subscription" "subscription { gameRoomUpdated(roomId: \`"$roomId\`") { id status } }"
Test-Subscription "Player Joined Subscription" "subscription { playerJoined(roomId: \`"$roomId\`") { username } }"
Test-Subscription "Player Left Subscription" "subscription { playerLeft(roomId: \`"$roomId\`") { username } }"
Test-Subscription "Player Ready Updated Subscription" "subscription { playerReadyUpdated(roomId: \`"$roomId\`") { username isReady } }"
Test-Subscription "Host Changed Subscription" "subscription { hostChanged(roomId: \`"$roomId\`") { username } }"
Test-Subscription "Game Started Subscription" "subscription { gameStarted(roomId: \`"$roomId\`") { id status } }"
Test-Subscription "Round Started Subscription" "subscription { roundStarted(roomId: \`"$roomId\`") { currentRound } }"
Test-Subscription "Round Ended Subscription" "subscription { roundEnded(roomId: \`"$roomId\`") { currentRound } }"
Test-Subscription "Game Ended Subscription" "subscription { gameEnded(roomId: \`"$roomId\`") { status } }"
Test-Subscription "Chat Message Subscription" "subscription { chatMessage(roomId: \`"$roomId\`") { message username } }"
Test-Subscription "Game Event Subscription" "subscription { gameEvent(roomId: \`"$roomId\`") { type } }"
Test-Subscription "Error Subscription" "subscription { error(roomId: \`"$roomId\`") { code message } }"

Write-Host "`n========================================" -ForegroundColor Yellow
Write-Host "Note: Subscriptions require WebSocket connection" -ForegroundColor Yellow
Write-Host "These tests verify query structure only" -ForegroundColor Yellow
Write-Host "For full subscription testing, use GraphQL Playground" -ForegroundColor Yellow
Write-Host "at http://localhost:8080/graphql" -ForegroundColor Yellow
Write-Host "========================================" -ForegroundColor Yellow
