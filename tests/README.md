# Tests

This directory contains integration and end-to-end test scripts for the Keycloak integration.

## Available Test Scripts

### final_test.sh
Comprehensive integration test that:
- Tests JWT authentication flow
- Validates hybrid middleware functionality
- Checks authorization and error handling
- Tests multiple endpoints
- Provides detailed test results

### test_keycloak.sh
Extended Keycloak-specific tests including:
- Keycloak server connectivity checks
- JWKS endpoint validation
- Token format verification
- Real Keycloak token testing (when server available)

## Running Tests

### Prerequisites
- Running application server (main app or test server)
- `curl` and `jq` installed
- (Optional) Running Keycloak server for full tests

### Quick Test
```bash
# Start test server first
go run simple_server.go &

# Run comprehensive test
./tests/final_test.sh
```

### Extended Keycloak Test
```bash
# Run Keycloak-specific tests
./tests/test_keycloak.sh
```

## Test Results

### Successful Test Output
```
🎯 Final Keycloak Integration Test
==================================

Testing: Health check
✅ SUCCESS (HTTP 200)

Getting JWT token...
✅ JWT token obtained

Testing: JWT token - Movies
✅ SUCCESS (HTTP 200)

=== Test Summary ===
✅ JWT Authentication: Working
✅ Hybrid Middleware: Functional
✅ Authorization: Working
⚠️  Keycloak: Ready (needs real server for full test)
```

## Test Coverage

These tests verify:
- ✅ JWT token generation and validation
- ✅ Hybrid middleware (JWT + Keycloak support)
- ✅ Authorization and access control
- ✅ Multiple endpoint protection
- ✅ Error handling and edge cases
- ✅ Configuration management

## CI/CD Integration

These scripts can be integrated into CI/CD pipelines:

```yaml
# Example GitHub Actions step
- name: Run Integration Tests
  run: |
    go run simple_server.go &
    sleep 2
    ./tests/final_test.sh
    pkill -f simple_server.go
```

## Troubleshooting

### Common Issues
1. **Server not running**: Ensure test server is started
2. **Port conflicts**: Check if port 8080 is available
3. **Missing dependencies**: Install `curl` and `jq`
4. **Permission denied**: Make scripts executable with `chmod +x`