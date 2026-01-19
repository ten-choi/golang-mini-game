# Sync mini-game-server to company q-connect-server
# 파일 동기화 스크립트

$sourceDir = "C:\workSpace\projects\personal\golang\golang-mini-game\mini-game-server"
$targetDir = "C:\workSpace\projects\company\q-connect-server"

Write-Host "=====================================" -ForegroundColor Cyan
Write-Host "  파일 동기화 시작" -ForegroundColor Cyan
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "소스: $sourceDir" -ForegroundColor Yellow
Write-Host "타겟: $targetDir" -ForegroundColor Yellow
Write-Host ""

# 소스 디렉토리 존재 확인
if (-not (Test-Path $sourceDir)) {
    Write-Host "❌ 소스 디렉토리가 존재하지 않습니다: $sourceDir" -ForegroundColor Red
    exit 1
}

# 타겟 디렉토리 존재 확인 및 생성
if (-not (Test-Path $targetDir)) {
    Write-Host "⚠️  타겟 디렉토리가 없습니다. 생성합니다: $targetDir" -ForegroundColor Yellow
    New-Item -ItemType Directory -Path $targetDir -Force | Out-Null
}

# 확인 메시지
Write-Host "⚠️  경고: 타겟 디렉토리의 기존 파일들이 덮어씌워집니다!" -ForegroundColor Red
Write-Host "계속하시겠습니까? (Y/N): " -NoNewline -ForegroundColor Yellow
$confirmation = Read-Host

if ($confirmation -ne 'Y' -and $confirmation -ne 'y') {
    Write-Host "작업이 취소되었습니다." -ForegroundColor Yellow
    exit 0
}

Write-Host ""
Write-Host "📂 파일 복사 중..." -ForegroundColor Green

try {
    # 모든 파일과 폴더 복사 (덮어쓰기)
    Copy-Item -Path "$sourceDir\*" -Destination $targetDir -Recurse -Force -ErrorAction Stop
    
    Write-Host ""
    Write-Host "✅ 동기화 완료!" -ForegroundColor Green
    Write-Host ""
    Write-Host "복사된 위치: $targetDir" -ForegroundColor Cyan
    Write-Host ""
    
    # 복사된 파일 개수 표시
    $fileCount = (Get-ChildItem -Path $targetDir -Recurse -File).Count
    $folderCount = (Get-ChildItem -Path $targetDir -Recurse -Directory).Count
    Write-Host "📊 통계:" -ForegroundColor Cyan
    Write-Host "  - 파일: $fileCount 개" -ForegroundColor White
    Write-Host "  - 폴더: $folderCount 개" -ForegroundColor White
    
} catch {
    Write-Host ""
    Write-Host "❌ 오류 발생: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "=====================================" -ForegroundColor Cyan
