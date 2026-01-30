# Sync mini-game-server to company q-connect-server
# File synchronization script

$sourceDir = "C:\workSpace\projects\personal\golang\golang-mini-game\mini-game-server"
$targetDir = "C:\workSpace\projects\company\q-connect-server"

Write-Host "=====================================" -ForegroundColor Cyan
Write-Host "  File Synchronization Started" -ForegroundColor Cyan
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Source: $sourceDir" -ForegroundColor Yellow
Write-Host "Target: $targetDir" -ForegroundColor Yellow
Write-Host ""

# Check source directory exists
if (-not (Test-Path $sourceDir)) {
    Write-Host "Source directory does not exist: $sourceDir" -ForegroundColor Red
    exit 1
}

# Check target directory exists and create if needed
if (-not (Test-Path $targetDir)) {
    Write-Host "Target directory does not exist. Creating: $targetDir" -ForegroundColor Yellow
    New-Item -ItemType Directory -Path $targetDir -Force | Out-Null
}

# Confirmation message
Write-Host "WARNING: Existing files in target directory will be overwritten!" -ForegroundColor Red
Write-Host "Do you want to continue? (Y/N): " -NoNewline -ForegroundColor Yellow
$confirmation = Read-Host

if ($confirmation -ne 'Y' -and $confirmation -ne 'y') {
    Write-Host "Operation cancelled." -ForegroundColor Yellow
    exit 0
}

Write-Host ""
Write-Host "Deleting existing files (excluding .git)..." -ForegroundColor Yellow

try {
    # Delete all items in target directory except .git
    Get-ChildItem -Path $targetDir -Exclude ".git" | Remove-Item -Recurse -Force -ErrorAction Stop
    
    Write-Host "Existing files deleted successfully" -ForegroundColor Green
    Write-Host ""
    Write-Host "Copying files..." -ForegroundColor Green
    
    # Copy all files and folders
    Copy-Item -Path "$sourceDir\*" -Destination $targetDir -Recurse -Force -ErrorAction Stop
    
    Write-Host ""
    Write-Host "Synchronization completed!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Copied to: $targetDir" -ForegroundColor Cyan
    Write-Host ""
    
    # Display file count statistics
    $fileCount = (Get-ChildItem -Path $targetDir -Recurse -File).Count
    $folderCount = (Get-ChildItem -Path $targetDir -Recurse -Directory).Count
    Write-Host "Statistics:" -ForegroundColor Cyan
    Write-Host "  - Files: $fileCount" -ForegroundColor White
    Write-Host "  - Folders: $folderCount" -ForegroundColor White
    
} catch {
    Write-Host ""
    Write-Host "Error occurred: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "=====================================" -ForegroundColor Cyan
