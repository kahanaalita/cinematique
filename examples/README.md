# Examples

This directory contains practical examples demonstrating key features of the cinematique project.

## Available Examples

### keycloak_example.go
Comprehensive example showing:
- How to create and initialize Keycloak clients
- Token validation with different options
- Using the Manager pattern for multiple clients
- Working with the Factory pattern
- Error handling best practices

### rate_limit_example.go
Rate limiting example demonstrating:
- Redis-based rate limiting implementation
- Fixed Window Counter algorithm
- Integration with Gin middleware
- Testing rate limits programmatically
- Error handling and monitoring

## Running Examples

### Prerequisites
- Go 1.22 or later
- (Optional) Running Keycloak server for full functionality

### Run the Keycloak example
```bash
# From project root
go run examples/keycloak_example.go
```

### Run the Rate Limiting example
```bash
# Make sure Redis is running
docker-compose up -d redis

# From project root
go run examples/rate_limit_example.go
```

### Expected Output
Without a Keycloak server:
```
This is not a Keycloak token
Failed to initialize manager: failed to initialize keycloak client: failed to fetch JWKS: ...
```

With a properly configured Keycloak server:
```
This is a Keycloak token
User ID: user-123
Username: testuser
Email: test@example.com
Roles: [user, cinematique-user]
Local Role: user
Keycloak is enabled
Default client is available
Client from factory: true
Total clients in factory: 1
```

## Integration with Main Application

These examples show the same patterns used in the main application:

1. **Configuration**: Same config structure as in `internal/config`
2. **Initialization**: Same patterns as in `cmd/app.go`
3. **Usage**: Same patterns as in `internal/auth/middleware.go`

## Customization

You can modify the examples to:
- Test with your own Keycloak server configuration
- Add custom validation options
- Experiment with different token formats
- Test error handling scenarios

## Best Practices Demonstrated

- ✅ Proper error handling
- ✅ Resource cleanup
- ✅ Configuration management
- ✅ Factory and Manager patterns
- ✅ Interface-based design