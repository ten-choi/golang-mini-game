$baseUrl = "http://localhost:8080/graphql"
$headers = @{ "Content-Type" = "application/json" }

function Invoke-GraphQL {
    param([string]$Query)
    $body = @{ query = $Query } | ConvertTo-Json -Depth 10
    try {
        $response = Invoke-RestMethod -Uri $baseUrl -Method Post -Headers $headers -Body $body
        return $response
    } catch {
        Write-Host "Error: $_" -ForegroundColor Red
        return $null
    }
}

Write-Host "`n=== Step 1: Create 4 users ===" -ForegroundColor Cyan
$users = @("player1", "player2", "player3", "player4")
foreach ($username in $users) {
    Write-Host "`nCreating: $username" -ForegroundColor Yellow
    $query = "mutation { createUser(input: {username: `"$username`", displayName: `"$username`"}) { id username displayName } }"
    $result = Invoke-GraphQL -Query $query
    if ($result.data.createUser) {
        Write-Host "OK - User created: $username (ID: $($result.data.createUser.id))" -ForegroundColor Green
    } else {
        Write-Host "FAIL - User creation failed: $($result.errors[0].message)" -ForegroundColor Red
    }
}

Write-Host "`n`n=== Step 2: player1 creates QA game room ===" -ForegroundColor Cyan
$createRoomQuery = 'mutation { createGameRoom(input: {name: "Test QA Game", gameType: QA, maxPlayers: 4, totalRounds: 3, hostUsername: "player1"}) { id name gameType status players { username } } }'
$roomResult = Invoke-GraphQL -Query $createRoomQuery
if ($roomResult.data.createGameRoom) {
    $roomId = $roomResult.data.createGameRoom.id
    Write-Host "OK - Game room created (ID: $roomId)" -ForegroundColor Green
    Write-Host "Room name: $($roomResult.data.createGameRoom.name)" -ForegroundColor Gray
    Write-Host "Game type: $($roomResult.data.createGameRoom.gameType)" -ForegroundColor Gray
} else {
    Write-Host "FAIL - Room creation failed: $($roomResult.errors[0].message)" -ForegroundColor Red
    exit 1
}

Write-Host "`n`n=== Step 3: 3 players join the room ===" -ForegroundColor Cyan
$joinPlayers = @("player2", "player3", "player4")
foreach ($player in $joinPlayers) {
    Write-Host "`nJoining: $player" -ForegroundColor Yellow
    $joinQuery = "mutation { joinGameRoom(roomId: `"$roomId`", username: `"$player`") { id players { username } } }"
    $joinResult = Invoke-GraphQL -Query $joinQuery
    if ($joinResult.data.joinGameRoom) {
        Write-Host "OK - $player joined (Total: $($joinResult.data.joinGameRoom.players.Count) players)" -ForegroundColor Green
    } else {
        Write-Host "FAIL - Join failed: $($joinResult.errors[0].message)" -ForegroundColor Red
    }
}

Write-Host "`n`n=== Step 4: Check room status ===" -ForegroundColor Cyan
$checkRoomQuery = "query { gameRoom(id: `"$roomId`") { id name status players { username score } currentRound totalRounds } }"
$checkResult = Invoke-GraphQL -Query $checkRoomQuery
if ($checkResult.data.gameRoom) {
    $room = $checkResult.data.gameRoom
    Write-Host "Room status: $($room.status)" -ForegroundColor Gray
    Write-Host "Players:" -ForegroundColor Gray
    foreach ($p in $room.players) {
        Write-Host "  - $($p.username) (Score: $($p.score))" -ForegroundColor Gray
    }
}

Write-Host "`n`n=== Step 5: Start game ===" -ForegroundColor Cyan
$startGameQuery = "mutation { startGame(roomId: `"$roomId`") { id status currentRound } }"
$startResult = Invoke-GraphQL -Query $startGameQuery
if ($startResult.data.startGame) {
    Write-Host "OK - Game started" -ForegroundColor Green
    Write-Host "Status: $($startResult.data.startGame.status)" -ForegroundColor Gray
} else {
    Write-Host "FAIL - Game start failed: $($startResult.errors[0].message)" -ForegroundColor Red
}

Write-Host "`n`n=== Step 6: Get random QA quiz ===" -ForegroundColor Cyan
$quizQuery = "query { randomQAQuiz(roomId: `"$roomId`") { id question options answer category } }"
$quizResult = Invoke-GraphQL -Query $quizQuery
if ($quizResult.data.randomQAQuiz) {
    $quiz = $quizResult.data.randomQAQuiz
    Write-Host "OK - Quiz loaded" -ForegroundColor Green
    Write-Host "Question: $($quiz.question)" -ForegroundColor Gray
    Write-Host "Category: $($quiz.category)" -ForegroundColor Gray
    if ($quiz.options) {
        Write-Host "Options:" -ForegroundColor Gray
        for ($i = 0; $i -lt $quiz.options.Count; $i++) {
            Write-Host "  $i. $($quiz.options[$i])" -ForegroundColor Gray
        }
        Write-Host "Correct answer index: $($quiz.answer)" -ForegroundColor Gray
    }
} else {
    Write-Host "FAIL - Quiz load failed: $($quizResult.errors[0].message)" -ForegroundColor Red
}

Write-Host "`n`n=== Step 7: Start round ===" -ForegroundColor Cyan
$startRoundQuery = "mutation { startRound(roomId: `"$roomId`") { id currentRound status } }"
$roundResult = Invoke-GraphQL -Query $startRoundQuery
if ($roundResult.data.startRound) {
    Write-Host "OK - Round 1 started" -ForegroundColor Green
    Write-Host "Current round: $($roundResult.data.startRound.currentRound)" -ForegroundColor Gray
} else {
    Write-Host "FAIL - Round start failed: $($roundResult.errors[0].message)" -ForegroundColor Red
}

Write-Host "`n`n=== Step 8: Players submit answers ===" -ForegroundColor Cyan
foreach ($player in $users) {
    $answer = Get-Random -Minimum 0 -Maximum 4
    Write-Host "`nSubmitting answer for $player (Answer: $answer)" -ForegroundColor Yellow
    $answerQuery = "mutation { submitAnswer(roomId: `"$roomId`", username: `"$player`", answer: `"$answer`") }"
    $answerResult = Invoke-GraphQL -Query $answerQuery
    if ($answerResult.data.submitAnswer) {
        Write-Host "OK - Answer submitted" -ForegroundColor Green
    } else {
        Write-Host "FAIL - Submit failed: $($answerResult.errors[0].message)" -ForegroundColor Red
    }
}

Write-Host "`n`n=== Step 9: Test chat ===" -ForegroundColor Cyan
$chatMessages = @("Hello!", "Good game", "Next round!")
for ($i = 0; $i -lt 3; $i++) {
    $player = $users[$i]
    $message = $chatMessages[$i]
    $chatQuery = "mutation { sendChat(roomId: `"$roomId`", username: `"$player`", message: `"$message`") { id username message } }"
    $chatResult = Invoke-GraphQL -Query $chatQuery
    if ($chatResult.data.sendChat) {
        Write-Host "OK - $player : $message" -ForegroundColor Green
    }
}

Write-Host "`n`n=== Step 10: End game ===" -ForegroundColor Cyan
$endGameQuery = "mutation { endGame(roomId: `"$roomId`") { id status } }"
$endResult = Invoke-GraphQL -Query $endGameQuery
if ($endResult.data.endGame) {
    Write-Host "OK - Game ended" -ForegroundColor Green
    Write-Host "Final status: $($endResult.data.endGame.status)" -ForegroundColor Gray
} else {
    Write-Host "FAIL - End game failed: $($endResult.errors[0].message)" -ForegroundColor Red
}

Write-Host "`n`n=== Step 11: Final room status ===" -ForegroundColor Cyan
$finalCheckQuery = "query { gameRoom(id: `"$roomId`") { id name status currentRound totalRounds players { username score } } }"
$finalResult = Invoke-GraphQL -Query $finalCheckQuery
if ($finalResult.data.gameRoom) {
    $finalRoom = $finalResult.data.gameRoom
    Write-Host "Final game room info:" -ForegroundColor Yellow
    Write-Host "Room name: $($finalRoom.name)" -ForegroundColor Gray
    Write-Host "Status: $($finalRoom.status)" -ForegroundColor Gray
    Write-Host "Rounds: $($finalRoom.currentRound)/$($finalRoom.totalRounds)" -ForegroundColor Gray
    Write-Host "`nFinal scores:" -ForegroundColor Yellow
    foreach ($p in $finalRoom.players) {
        Write-Host "  $($p.username): $($p.score) points" -ForegroundColor Gray
    }
}

Write-Host "`n`n=== TEST COMPLETED! ===" -ForegroundColor Green
