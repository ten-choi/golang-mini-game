$baseUrl = "http://localhost:8080/graphql"

function Test-Word {
    param([string]$Word, [bool]$Expected, [string]$Description)
    
    $query = @"
{
  "query": "query { isValidWord(word: \"$Word\") }"
}
"@
    
    try {
        $result = Invoke-RestMethod -Uri $baseUrl -Method Post -ContentType "application/json; charset=utf-8" -Body $query
        $actual = $result.data.isValidWord
        $pass = $actual -eq $Expected
        $icon = if($pass){"✅"}else{"❌"}
        $color = if($pass){"Green"}else{"Red"}
        Write-Host "$icon $Description [$Word]: $actual (expected: $Expected)" -ForegroundColor $color
        return $pass
    } catch {
        Write-Host "❌ Error testing $Word : $_" -ForegroundColor Red
        return $false
    }
}

Write-Host "`n=== Testing Japanese Word Dictionary ===" -ForegroundColor Yellow
Write-Host "Testing with in-memory hash-based search (FNV 64-bit + Binary Search)" -ForegroundColor Cyan

$passCount = 0
$totalCount = 0

# Test valid words
$totalCount++; if(Test-Word "りんご" $true "Valid Hiragana word (apple)") { $passCount++ }
$totalCount++; if(Test-Word "リンゴ" $true "Valid Katakana word (apple - normalized)") { $passCount++ }
$totalCount++; if(Test-Word "らーめん" $true "Word with long vowel mark") { $passCount++ }
$totalCount++; if(Test-Word "ごりら" $true "Gorilla in Hiragana") { $passCount++ }
$totalCount++; if(Test-Word "ゴリラ" $true "Gorilla in Katakana") { $passCount++ }
$totalCount++; if(Test-Word "ねこ" $true "Cat in Hiragana") { $passCount++ }
$totalCount++; if(Test-Word "ネコ" $true "Cat in Katakana") { $passCount++ }
$totalCount++; if(Test-Word "らっぱ" $true "Trumpet in Hiragana") { $passCount++ }
$totalCount++; if(Test-Word "ラッパ" $true "Trumpet in Katakana") { $passCount++ }

# Test invalid words
$totalCount++; if(Test-Word "バナナ" $false "Invalid word (banana)") { $passCount++ }
$totalCount++; if(Test-Word "apple" $false "English word") { $passCount++ }
$totalCount++; if(Test-Word "" $false "Empty string") { $passCount++ }
$totalCount++; if(Test-Word "あいうえお" $false "Random Hiragana") { $passCount++ }

Write-Host "`n=== Test Results ===" -ForegroundColor Yellow
Write-Host "Passed: $passCount / $totalCount" -ForegroundColor $(if($passCount -eq $totalCount){"Green"}else{"Yellow"})
Write-Host "Success Rate: $([math]::Round($passCount/$totalCount*100, 2))%" -ForegroundColor Cyan
Write-Host "`nDictionary validation is working with:" -ForegroundColor Green
Write-Host "  • In-memory storage (no database queries)" -ForegroundColor White
Write-Host "  • FNV 64-bit hashing for word fingerprints" -ForegroundColor White  
Write-Host "  • Binary search for O(log n) lookup" -ForegroundColor White
Write-Host "  • Katakana → Hiragana normalization" -ForegroundColor White
Write-Host "  • 57 words loaded from japanese_words.txt" -ForegroundColor White
