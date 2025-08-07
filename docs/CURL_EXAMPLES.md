# CURL Examples for Cinematique API

This document provides practical CURL examples for testing the Cinematique API endpoints, including rate limiting features.

## Authentication

### Register a new user
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "testpass123",
    "role": "user"
  }'
```

### Login
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "testpass123"
  }'
```

Response:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

### Refresh Token
```bash
curl -X POST http://localhost:8080/api/auth/refresh \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Logout
```bash
curl -X POST http://localhost:8080/api/auth/logout \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## Rate Limiting

### Check rate limit status
```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  "http://localhost:8080/api/rate-limit/status?endpoint=/api/movies"
```

Response:
```json
{
  "enabled": true,
  "user_id": "testuser",
  "ip": "127.0.0.1",
  "endpoint": "/api/movies",
  "current_count": 45,
  "limit": 1000,
  "remaining": 955,
  "window": "1m0s",
  "reset_time": 1706097600,
  "reset_time_human": "2025-01-24T10:05:00Z",
  "restricted_endpoints": ["/api/movies", "/api/actors"]
}
```

### Test rate limiting (make many requests)
```bash
# This will eventually trigger rate limiting
for i in {1..1005}; do
  curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
    http://localhost:8080/api/movies
  sleep 0.01
done
```

### Rate limit exceeded response (429)
```json
{
  "error": "Too many requests",
  "message": "Rate limit exceeded. Maximum 1000 requests per 1m0s",
  "current_count": 1001,
  "limit": 1000,
  "window": "1m0s",
  "retry_after": 60
}
```

## Movies

### Get all movies (with rate limit headers)
```bash
curl -I -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/api/movies
```

Response headers include:
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 995
X-RateLimit-Reset: 1706097600
```

### Get movie by ID
```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/api/movies/1
```

### Search movies by title
```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  "http://localhost:8080/api/movies/search?title=Inception"
```

### Search movies by actor name
```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  "http://localhost:8080/api/movies/search?actorName=Leonardo"
```

### Get sorted movies
```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  "http://localhost:8080/api/movies/sorted?sort=title&order=asc"
```

### Create a new movie (Admin only)
```bash
curl -X POST http://localhost:8080/api/movies \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_JWT_TOKEN" \
  -d '{
    "title": "Inception",
    "description": "A mind-bending thriller",
    "release_date": "2010-07-16",
    "rating": 8.8
  }'
```

### Update movie (Admin only)
```bash
curl -X PUT http://localhost:8080/api/movies/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_JWT_TOKEN" \
  -d '{
    "title": "Inception Updated",
    "description": "An updated mind-bending thriller",
    "release_date": "2010-07-16",
    "rating": 9.0
  }'
```

### Partial update movie (Admin only)
```bash
curl -X PATCH http://localhost:8080/api/movies/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_JWT_TOKEN" \
  -d '{
    "rating": 9.2
  }'
```

### Delete movie (Admin only)
```bash
curl -X DELETE http://localhost:8080/api/movies/1 \
  -H "Authorization: Bearer YOUR_ADMIN_JWT_TOKEN"
```

## Actors

### Get all actors
```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/api/actors
```

### Get actors with their movies
```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/api/actors/with-movies
```

### Get actor by ID
```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/api/actors/1
```

### Create a new actor (Admin only)
```bash
curl -X POST http://localhost:8080/api/actors \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_JWT_TOKEN" \
  -d '{
    "name": "Leonardo DiCaprio",
    "gender": "male",
    "birth_date": "1974-11-11"
  }'
```

### Update actor (Admin only)
```bash
curl -X PUT http://localhost:8080/api/actors/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_JWT_TOKEN" \
  -d '{
    "name": "Leonardo Wilhelm DiCaprio",
    "gender": "male",
    "birth_date": "1974-11-11"
  }'
```

### Partial update actor (Admin only)
```bash
curl -X PATCH http://localhost:8080/api/actors/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_JWT_TOKEN" \
  -d '{
    "name": "Leo DiCaprio"
  }'
```

### Delete actor (Admin only)
```bash
curl -X DELETE http://localhost:8080/api/actors/1 \
  -H "Authorization: Bearer YOUR_ADMIN_JWT_TOKEN"
```

## Movie-Actor Relationships

### Get actors for a movie
```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/api/movies/1/actors
```

### Get movies for an actor
```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/api/movies/actor/1
```

### Add actor to movie (Admin only)
```bash
curl -X POST http://localhost:8080/api/movies/add-actor/1/2 \
  -H "Authorization: Bearer YOUR_ADMIN_JWT_TOKEN"
```

### Remove actor from movie (Admin only)
```bash
curl -X DELETE http://localhost:8080/api/movies/remove-actor/1/2 \
  -H "Authorization: Bearer YOUR_ADMIN_JWT_TOKEN"
```

### Create movie with actors (Admin only)
```bash
curl -X POST http://localhost:8080/api/movies/with-actors \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_JWT_TOKEN" \
  -d '{
    "title": "The Dark Knight",
    "description": "Batman fights the Joker",
    "release_date": "2008-07-18",
    "rating": 9.0,
    "actor_ids": [1, 2, 3]
  }'
```

### Update movie actors (Admin only)
```bash
curl -X POST http://localhost:8080/api/movies/1/actors \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_JWT_TOKEN" \
  -d '{
    "actor_ids": [1, 3, 4]
  }'
```

## Monitoring

### Prometheus metrics
```bash
curl http://localhost:8080/metrics
```

## Rate Limiting Testing

### Test different users (different limits)
```bash
# User 1
curl -H "Authorization: Bearer USER1_JWT_TOKEN" \
  http://localhost:8080/api/movies

# User 2 (separate rate limit)
curl -H "Authorization: Bearer USER2_JWT_TOKEN" \
  http://localhost:8080/api/movies
```

### Test different IPs (using proxy headers)
```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "X-Forwarded-For: 192.168.1.100" \
  http://localhost:8080/api/movies

curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "X-Forwarded-For: 192.168.1.200" \
  http://localhost:8080/api/movies
```

### Test different endpoints
```bash
# Movies endpoint
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/api/movies

# Actors endpoint (separate rate limit)
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/api/actors
```

## Error Handling

### Invalid token
```bash
curl -H "Authorization: Bearer invalid_token" \
  http://localhost:8080/api/movies
```

Response:
```json
{
  "error": "Invalid token"
}
```

### Rate limit exceeded (429)
```bash
# After making too many requests
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/api/movies
```

Response:
```json
{
  "error": "Too many requests",
  "message": "Rate limit exceeded. Maximum 1000 requests per 1m0s",
  "current_count": 1001,
  "limit": 1000,
  "window": "1m0s",
  "retry_after": 60
}
```

### Insufficient permissions
```bash
curl -X POST http://localhost:8080/api/movies \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer USER_JWT_TOKEN" \
  -d '{
    "title": "Test Movie",
    "description": "Test",
    "release_date": "2023-01-01",
    "rating": 7.0
  }'
```

Response:
```json
{
  "error": "Insufficient permissions"
}
```

## Testing Scripts

For automated testing, use the provided scripts:

```bash
# Test rate limiting specifically
./tests/test_rate_limit.sh

# Test authentication
./tests/test_keycloak.sh

# Test all functionality including rate limiting
./tests/final_test.sh
```

## Notes

- Replace `YOUR_JWT_TOKEN` with the actual token from login response
- Replace `YOUR_ADMIN_JWT_TOKEN` with a token from an admin user
- All timestamps are in RFC3339 format
- The API uses JSON for request and response bodies
- Authentication is required for all endpoints except `/auth/*`
- Rate limiting applies to `/api/movies` and `/api/actors` endpoints by default
- Rate limit headers are included in all responses from protected endpoints
- Each user+IP+endpoint combination has its own rate limit counter