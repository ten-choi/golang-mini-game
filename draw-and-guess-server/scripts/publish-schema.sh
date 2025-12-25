#!/usr/bin/env bash

# Apollo Studio Schema Registry Upload Script
# This script uploads the GraphQL schema to Apollo Studio for documentation

set -e

# Configuration
APOLLO_KEY="${APOLLO_KEY:-}"
APOLLO_GRAPH_REF="${APOLLO_GRAPH_REF:-draw-and-guess@main}"
SCHEMA_FILE="src/graph/schema-complete.graphqls"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🚀 Apollo Schema Registry Upload${NC}"
echo "=================================="

# Check if rover is installed
if ! command -v rover &> /dev/null; then
    echo -e "${RED}❌ Rover CLI is not installed${NC}"
    echo ""
    echo "Please install Rover CLI:"
    echo "  npm install -g @apollo/rover"
    echo "  or"
    echo "  curl -sSL https://rover.apollo.dev/nix/latest | sh"
    echo ""
    exit 1
fi

# Check if APOLLO_KEY is set
if [ -z "$APOLLO_KEY" ]; then
    echo -e "${RED}❌ APOLLO_KEY environment variable is not set${NC}"
    echo ""
    echo "Please set your Apollo Studio API key:"
    echo "  export APOLLO_KEY='your-apollo-key'"
    echo ""
    echo "Get your API key from: https://studio.apollographql.com/user-settings/api-keys"
    echo ""
    exit 1
fi

# Check if schema file exists
if [ ! -f "$SCHEMA_FILE" ]; then
    echo -e "${RED}❌ Schema file not found: $SCHEMA_FILE${NC}"
    exit 1
fi

echo -e "${YELLOW}📋 Validating schema...${NC}"
rover graph check "$APOLLO_GRAPH_REF" \
    --schema "$SCHEMA_FILE" \
    || echo -e "${YELLOW}⚠️  Schema check warnings (non-blocking)${NC}"

echo ""
echo -e "${YELLOW}📤 Publishing schema to Apollo Studio...${NC}"
rover graph publish "$APOLLO_GRAPH_REF" \
    --schema "$SCHEMA_FILE" \
    --routing-url "http://localhost:8080/graphql"

echo ""
echo -e "${GREEN}✅ Schema successfully published!${NC}"
echo ""
echo "View your schema at:"
echo -e "${GREEN}https://studio.apollographql.com/graph/${APOLLO_GRAPH_REF%%@*}${NC}"
echo ""
echo "API Documentation is now available for:"
echo "  • GraphQL Operations"
echo "  • REST API Endpoints (documented in schema comments)"
echo "  • WebSocket Protocol (documented in schema comments)"
