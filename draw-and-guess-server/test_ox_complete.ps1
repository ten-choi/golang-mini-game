# Complete OX Quiz Test with Logging

Write-Host "`n========== OX QUIZ 완전 테스트 ==========`n" -ForegroundColor Cyan

# Create OX Room
Write-Host "[1/4] OX 게임방 생성..." -ForegroundColor Yellow
$query1 = '{"query":"mutation { createGameRoom(input: { name: \"OX Test\", gameType: OX, maxPlayers: 4, roundTimeLimit: 15, totalRounds: 3, hostUsername: \"tester\" }) { id name gameType status } }"}'
$resp1 = Invoke-RestMethod -Uri "http://localhost:8080/graphql" -Method POST -Body $query1 -ContentType "application/json"
$roomId = $resp1.data.createGameRoom.id
Write-Host "✅ 방 생성 완료: $roomId`n" -ForegroundColor Green

# Join Room
Write-Host "[2/4] 플레이어 참가..." -ForegroundColor Yellow
$query2 = "{`"query`":`"mutation { joinGameRoom(roomId: \`"$roomId\`", username: \`"tester2\`") { id players { username } } }`"}"
$resp2 = Invoke-RestMethod -Uri "http://localhost:8080/graphql" -Method POST -Body $query2 -ContentType "application/json"
Write-Host "✅ 플레이어 수: $($resp2.data.joinGameRoom.players.Count)`n" -ForegroundColor Green

# Start Game
Write-Host "[3/4] 게임 시작..." -ForegroundColor Yellow
$query3 = "{`"query`":`"mutation { startGame(roomId: \`"$roomId\`") { id status currentRound } }`"}"
$resp3 = Invoke-RestMethod -Uri "http://localhost:8080/graphql" -Method POST -Body $query3 -ContentType "application/json"
Write-Host "✅ 게임 상태: $($resp3.data.startGame.status)" -ForegroundColor Green
Write-Host "✅ 현재 라운드: $($resp3.data.startGame.currentRound)`n" -ForegroundColor Green

# Wait and check logs
Write-Host "[4/4] 로그 확인 (3초 대기)..." -ForegroundColor Yellow
Start-Sleep -Seconds 3

# Try to find server log window
$serverWindows = Get-Process powershell | Where-Object {$_.MainWindowTitle -like "*server.exe*"}
if ($serverWindows) {
    Write-Host "`n💡 서버 로그 창을 확인하세요!" -ForegroundColor Cyan
}

Write-Host "`n========== 테스트 완료 ==========`n" -ForegroundColor Cyan
Write-Host "다음을 확인하세요:" -ForegroundColor Yellow
Write-Host "  1. 서버 콘솔에 [Quiz] 로그가 출력되었는지" -ForegroundColor White
Write-Host "  2. [Quiz] SUCCESS: Got OX quiz 메시지가 있는지" -ForegroundColor White
Write-Host "  3. 브라우저(http://localhost:5174)에서 방 $roomId 입장" -ForegroundColor White
Write-Host "  4. OX 퀴즈 문제와 O/X 버튼이 보이는지 확인" -ForegroundColor White
Write-Host ""
