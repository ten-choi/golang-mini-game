# Response Handling Guide

## REST vs GraphQL

### REST Endpoints (Use common.Response functions)
```go
// Success response (200)
common.SuccessResponse(c, userData)
// Output: {"id": 1, "name": "user"}

// Created response (201)
common.CreatedResponse(c, newUser)

// Error with validation details
details := map[string]string{
    "email": "invalid format",
    "username": "already exists",
}
common.ErrorResponseWithDetails(c, err, details)
// Output: 
// {
//   "error": {
//     "code": 400,
//     "message": "Validation failed",
//     "details": {"email": "invalid format"},
//     "trace_id": "20251225-a1b2c3d4"
//   }
// }
```

### GraphQL Resolvers (Return data directly)
```go
func (r *queryResolver) User(ctx context.Context, username string) (*model.User, error) {
    // DO NOT use common.SuccessResponse here
    // Just return data and error
    user, err := r.UserService.GetByUsername(username)
    if err != nil {
        return nil, err  // gqlgen handles error formatting
    }
    return user, nil
}

// GraphQL automatically wraps:
// Success: {"data": {"user": {"id": 1, "name": "user"}}}
// Error: {"errors": [{"message": "Not Found", "path": ["user"]}]}
```

## Error Response Format

### Standard Error
```json
{
  "error": {
    "code": 404,
    "message": "User not found",
    "trace_id": "20251225150405-a1b2c3d4"
  }
}
```

### Validation Error with Details
```json
{
  "error": {
    "code": 400,
    "message": "Validation failed",
    "details": {
      "email": "Invalid email format",
      "password": "Must be at least 8 characters"
    },
    "trace_id": "20251225150405-a1b2c3d4"
  }
}
```

## Trace ID

Every request gets a unique trace ID:
- **Auto-generated** by middleware
- **Included in response** header: `X-Trace-ID`
- **Included in error response** body: `trace_id`
- **Used for log correlation** to track request flow

Example:
```bash
curl -v http://localhost:8080/api/v1/users
< X-Trace-ID: 20251225150405-a1b2c3d4
```

Use this trace ID to search logs when debugging issues.

## Best Practices

### ✅ Do
- Use `SuccessResponse` for REST endpoints
- Return data directly in GraphQL resolvers
- Include validation details with `ErrorResponseWithDetails`
- Check trace ID in logs when debugging

### ❌ Don't
- Don't use common.Response functions in GraphQL resolvers
- Don't manually add `success` field (unenveloped)
- Don't ignore the trace ID in error responses
- Don't use c.JSON directly (prefer helper functions for consistency)

## REST Endpoint Example

```go
func CreateUser(c *gin.Context) {
    var input CreateUserInput
    if err := c.ShouldBindJSON(&input); err != nil {
        details := map[string]string{
            "input": "Invalid JSON format",
        }
        common.ErrorResponseWithDetails(c, 
            common.NewBadRequestError("Invalid input", err), 
            details)
        return
    }
    
    user, err := userService.Create(input)
    if err != nil {
        common.ErrorResponseJSON(c, err)
        return
    }
    
    common.CreatedResponse(c, user)
}
```
