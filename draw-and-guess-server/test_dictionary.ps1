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

Write-Host "`n=== Testing JMdict Dictionary (211,029 words) ===" -ForegroundColor Yellow
Write-Host "With 2+ characters, no maximum length limit" -ForegroundColor Cyan

$passCount = 0
$totalCount = 0

# Test short words
$totalCount++; if(Test-Word "りんご" $true "Apple (4 chars)") { $passCount++ }
$totalCount++; if(Test-Word "リンゴ" $true "Apple in Katakana") { $passCount++ }
$totalCount++; if(Test-Word "ねこ" $true "Cat (2 chars)") { $passCount++ }
$totalCount++; if(Test-Word "ネコ" $true "Cat in Katakana") { $passCount++ }

# Test longer words (now should be valid!)
$totalCount++; if(Test-Word "バナナ" $true "Banana (3 chars)") { $passCount++ }
$totalCount++; if(Test-Word "コンピュータ" $true "Computer (7 chars)") { $passCount++ }
$totalCount++; if(Test-Word "おもしろい" $true "Interesting (5 chars)") { $passCount++ }
$totalCount++; if(Test-Word "プログラミング" $true "Programming (8 chars)") { $passCount++ }

# Test very long words
$totalCount++; if(Test-Word "インターネット" $true "Internet (8 chars)") { $passCount++ }
$totalCount++; if(Test-Word "スマートフォン" $true "Smartphone (8 chars)") { $passCount++ }

# Test invalid words
$totalCount++; if(Test-Word "あ" $false "Single character (too short)") { $passCount++ }
$totalCount++; if(Test-Word "apple" $false "English word") { $passCount++ }
$totalCount++; if(Test-Word "" $false "Empty string") { $passCount++ }
$totalCount++; if(Test-Word "あいうえお" $true "Random valid hiragana") { $passCount++ }

Write-Host "`n=== Test Results ===" -ForegroundColor Yellow
Write-Host "Passed: $passCount / $totalCount" -ForegroundColor $(if($passCount -eq $totalCount){"Green"}else{"Yellow"})
Write-Host "Success Rate: $([math]::Round($passCount/$totalCount*100, 2))%" -ForegroundColor Cyan
Write-Host "`nDictionary powered by JMdict:" -ForegroundColor Green
Write-Host "  • 211,029 words from JMdict-eng-3.6.1.json" -ForegroundColor White
Write-Host "  • Minimum 2 characters (no maximum limit)" -ForegroundColor White
Write-Host "  • In-memory FNV 64-bit hashing" -ForegroundColor White  
Write-Host "  • Binary search O(log n) lookup" -ForegroundColor White
Write-Host "  • Katakana → Hiragana normalization" -ForegroundColor White
