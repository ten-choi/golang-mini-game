# Simple OX Test
$roomId = ""

# Create room
Write-Host "Creating OX room..." -ForegroundColor Yellow
$createQuery = '{"query":"mutation { createGameRoom(input: { name: \"OX Test Room\", gameType: OX, maxPlayers: 4, roundTimeLimit: 15, totalRounds: 3, hostUsername: \"testuser\" }) { id name gameType } }"}'
$createResp = Invoke-RestMethod -Uri "http://localhost:8080/graphql" -Method POST -Body $createQuery -ContentType "application/json"
$roomId = $createResp.data.createGameRoom.id
Write-Host "✅ Room created: $roomId"

# Join room
Write-Host "Joining room..." -ForegroundColor Yellow
$joinQuery = "{`"query`":`"mutation { joinGameRoom(roomId: \`"$roomId\`", username: \`"testuser2\`") { id } }`"}"
Invoke-RestMethod -Uri "http://localhost:8080/graphql" -Method POST -Body $joinQuery -ContentType "application/json" | Out-Null
Write-Host "✅ Player joined"

# Start game
Write-Host "Starting game..." -ForegroundColor Yellow
$startQuery = "{`"query`":`"mutation { startGame(roomId: \`"$roomId\`") { id status currentRound } }`"}"
Invoke-RestMethod -Uri "http://localhost:8080/graphql" -Method POST -Body $startQuery -ContentType "application/json" | Out-Null
Write-Host "✅ Game started!"

Write-Host "`n⏳ Waiting 3 seconds for quiz to be sent..." -ForegroundColor Cyan
Start-Sleep -Seconds 3

Write-Host "`n========== Server Logs ==========`n" -ForegroundColor Cyan
Get-Content "server_all.log" -Tail 40 | Select-String "OX|Quiz|StartGame" | ForEach-Object {
    if ($_ -match "ERROR") {
        Write-Host $_ -ForegroundColor Red
    } elseif ($_ -match "SUCCESS") {
        Write-Host $_ -ForegroundColor Green
    } else {
        Write-Host $_
    }
}
