#!/bin/bash
set -e

echo "Testing ShopPilot API Authentication"

# Health check
echo "1. Health check (public)..."
curl -f http://localhost:8080/health || (echo "FAIL: Server not running" && exit 1)
echo "✓ Server is running"

# Login
echo "2. Login..."
RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@test.com","password":"password123"}')

TOKEN=$(echo $RESPONSE | grep -o '"token":"[^"]*' | cut -d'"' -f4)
if [ -z "$TOKEN" ]; then
  echo "FAIL: No token received"
  exit 1
fi
echo "✓ Login successful, token: ${TOKEN:0:20}..."

# Protected endpoint
echo "3. Access protected endpoint..."
curl -f -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/clients || (echo "FAIL: Protected route failed" && exit 1)
echo "✓ Protected route accessible"

echo ""
echo "✅ All authentication tests passed!"
