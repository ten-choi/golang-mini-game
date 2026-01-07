$baseUrl = "http://localhost:8080/graphql"

function Invoke-GraphQL {
    param([string]$Query)
    $body = @{ query = $Query } | ConvertTo-Json -Depth 10
    try {
        $result = Invoke-RestMethod -Uri $baseUrl -Method Post -ContentType "application/json" -Body $body -TimeoutSec 10
        return $result
    } catch {
        Write-Host "Error: $_" -ForegroundColor Red
        return $null
    }
}

Write-Host "`n=== Testing Password Protection ===" -ForegroundColor Yellow

# Create test user
Write-Host "`n1. Creating test users..." -ForegroundColor Cyan
$user1 = Invoke-GraphQL 'mutation { createUser(input: {username: "pwtest1", displayName: "PW Test 1"}) { id username } }'
$user2 = Invoke-GraphQL 'mutation { createUser(input: {username: "pwtest2", displayName: "PW Test 2"}) { id username } }'
Write-Host "Users created" -ForegroundColor Green

# Create private room with password
Write-Host "`n2. Creating private room with password '1234'..." -ForegroundColor Cyan
$createRoom = Invoke-GraphQL 'mutation { createGameRoom(input: {name: "Private Test Room", gameType: OX, maxPlayers: 4, totalRounds: 3, hostUsername: "pwtest1", isPrivate: true, password: "1234"}) { id name isPrivate password } }'
if ($createRoom.data.createGameRoom) {
    $roomId = $createRoom.data.createGameRoom.id
    Write-Host "Room created: $roomId" -ForegroundColor Green
    Write-Host "IsPrivate: $($createRoom.data.createGameRoom.isPrivate)" -ForegroundColor Gray
} else {
    Write-Host "Failed: $($createRoom.errors[0].message)" -ForegroundColor Red
    exit
}

# Test 1: Join without password
Write-Host "`n3. Test: Join without password (should FAIL)..." -ForegroundColor Cyan
$joinNoPassword = Invoke-GraphQL "mutation { joinGameRoom(roomId: `"$roomId`", username: `"pwtest2`") { id } }"
if ($joinNoPassword.errors) {
    Write-Host "✅ Correctly rejected (no password)" -ForegroundColor Green
    Write-Host "Error: $($joinNoPassword.errors[0].message)" -ForegroundColor Gray
} else {
    Write-Host "❌ FAILED: Should have rejected but succeeded" -ForegroundColor Red
}

# Test 2: Join with wrong password
Write-Host "`n4. Test: Join with wrong password (should FAIL)..." -ForegroundColor Cyan
$joinWrongPW = Invoke-GraphQL "mutation { joinGameRoom(roomId: `"$roomId`", username: `"pwtest2`", password: `"wrong`") { id } }"
if ($joinWrongPW.errors) {
    Write-Host "✅ Correctly rejected (wrong password)" -ForegroundColor Green
    Write-Host "Error: $($joinWrongPW.errors[0].message)" -ForegroundColor Gray
} else {
    Write-Host "❌ FAILED: Should have rejected but succeeded" -ForegroundColor Red
}

# Test 3: Join with correct password
Write-Host "`n5. Test: Join with correct password (should SUCCEED)..." -ForegroundColor Cyan
$joinCorrectPW = Invoke-GraphQL "mutation { joinGameRoom(roomId: `"$roomId`", username: `"pwtest2`", password: `"1234`") { id players { username } } }"
if ($joinCorrectPW.data.joinGameRoom) {
    Write-Host "✅ Successfully joined with correct password" -ForegroundColor Green
    Write-Host "Players: $($joinCorrectPW.data.joinGameRoom.players.Count)" -ForegroundColor Gray
} else {
    Write-Host "❌ FAILED: Should have succeeded" -ForegroundColor Red
    if ($joinCorrectPW.errors -and $joinCorrectPW.errors.Count -gt 0) {
        Write-Host "Error: $($joinCorrectPW.errors[0].message)" -ForegroundColor Gray
    }
}

Write-Host "`n=== Test Completed ===" -ForegroundColor Yellow
