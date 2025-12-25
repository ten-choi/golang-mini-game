# Apollo Studio Schema Registry Upload Script (PowerShell)
# This script uploads the GraphQL schema to Apollo Studio for documentation

param(
    [string]$ApolloKey = $env:APOLLO_KEY,
    [string]$GraphRef = $env:APOLLO_GRAPH_REF
)

# Default values
if (-not $GraphRef) {
    $GraphRef = "draw-and-guess@main"
}

$SchemaFile = "src\graph\schema-complete.graphqls"

Write-Host "🚀 Apollo Schema Registry Upload" -ForegroundColor Green
Write-Host "==================================" -ForegroundColor Green
Write-Host ""

# Check if rover is installed
$roverInstalled = Get-Command rover -ErrorAction SilentlyContinue
if (-not $roverInstalled) {
    Write-Host "❌ Rover CLI is not installed" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please install Rover CLI:"
    Write-Host "  npm install -g @apollo/rover"
    Write-Host ""
    exit 1
}

# Check if APOLLO_KEY is set
if (-not $ApolloKey) {
    Write-Host "❌ APOLLO_KEY environment variable is not set" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please set your Apollo Studio API key:"
    Write-Host "  `$env:APOLLO_KEY='your-apollo-key'"
    Write-Host ""
    Write-Host "Get your API key from: https://studio.apollographql.com/user-settings/api-keys"
    Write-Host ""
    exit 1
}

# Set environment variable for rover
$env:APOLLO_KEY = $ApolloKey

# Check if schema file exists
if (-not (Test-Path $SchemaFile)) {
    Write-Host "❌ Schema file not found: $SchemaFile" -ForegroundColor Red
    exit 1
}

Write-Host "📋 Validating schema..." -ForegroundColor Yellow
rover graph check $GraphRef --schema $SchemaFile
if ($LASTEXITCODE -ne 0) {
    Write-Host "⚠️  Schema check warnings (non-blocking)" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "📤 Publishing schema to Apollo Studio..." -ForegroundColor Yellow
rover graph publish $GraphRef --schema $SchemaFile --routing-url "http://localhost:8080/graphql"

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "✅ Schema successfully published!" -ForegroundColor Green
    Write-Host ""
    $graphName = $GraphRef.Split('@')[0]
    Write-Host "View your schema at:"
    Write-Host "https://studio.apollographql.com/graph/$graphName" -ForegroundColor Green
    Write-Host ""
    Write-Host "API Documentation is now available for:"
    Write-Host "  • GraphQL Operations"
    Write-Host "  • REST API Endpoints (documented in schema comments)"
    Write-Host "  • WebSocket Protocol (documented in schema comments)"
} else {
    Write-Host ""
    Write-Host "❌ Schema publication failed" -ForegroundColor Red
    exit 1
}
