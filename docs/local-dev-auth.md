# Local Development - Authentication

## Prerequisites
- PostgreSQL database running
- Redis running
- Environment variables configured

## Required Environment Variables
```bash
# JWT Configuration (add to .env)
JWT_SECRET=dev-secret-change-in-production
JWT_EXPIRATION_HOURS=24
```

## Running Locally
```bash
# Start the server
go run cmd/api/main.go

# Server should start on http://localhost:8080
```

## Testing Authentication

### 1. Create a test user (run SQL)
```sql
-- Insert test platform user
INSERT INTO platform_users (id, email, username, password, first_name, last_name, user_status_id)
VALUES (
  gen_random_uuid(),
  'admin@test.com',
  'admin',
  '$2a$10$X.UOmS7xGFBhSqLqJpG4GuZCZ8jtC8YpLVZ8H7vZvZqJxDqJxDqJx', -- "password123"
  'Test',
  'Admin',
  1
);
```

### 2. Get auth token
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@test.com","password":"password123"}'
```

### 3. Use token in requests
```bash
export TOKEN="your-token-here"
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/clients
```
