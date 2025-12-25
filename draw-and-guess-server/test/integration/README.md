# Integration Tests

This directory contains integration tests that test multiple components together.

## Running Tests

```bash
# Run all integration tests
go test ./test/integration/...

# Run with verbose output
go test -v ./test/integration/...

# Run specific test
go test -v ./test/integration -run TestHealthEndpoint
```

## Test Structure

- Each test file should focus on a specific feature or API endpoint
- Use test fixtures and setup/teardown functions
- Mock external dependencies (database, cache) when needed
