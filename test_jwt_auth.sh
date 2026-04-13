#!/bin/bash
# JWT Testing Script - Test authenticated endpoints with JWT tokens

API_URL="http://localhost:8080"

echo "🔐 JWT Authentication Testing Script"
echo "===================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Step 1: LOGIN
echo -e "${YELLOW}[1/5]${NC} Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "testpass123"
  }')

echo "Response: $LOGIN_RESPONSE"
echo ""

# Extract token
TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"Token":"[^"]*' | grep -o '[^"]*$')

if [ -z "$TOKEN" ]; then
    echo -e "${RED}❌ Failed to get token${NC}"
    echo "Make sure you have a test user in the database:"
    echo "  curl -X POST http://localhost:8080/users \\"
    echo "    -H 'Content-Type: application/json' \\"
    echo "    -d '{\"username\":\"testuser\",\"email\":\"test@example.com\",\"password\":\"testpass123\"}'"
    exit 1
fi

echo -e "${GREEN}✓ Token received:${NC}"
echo "  ${TOKEN:0:50}..."
echo ""

# Step 2: CREATE ESTATE
echo -e "${YELLOW}[2/5]${NC} Creating estate..."
ESTATE_RESPONSE=$(curl -s -X POST "$API_URL/estate" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "length": 100,
    "width": 100
  }')

echo "Response: $ESTATE_RESPONSE"
ESTATE_ID=$(echo $ESTATE_RESPONSE | grep -o '"Id":"[^"]*' | head -1 | grep -o '[^"]*$')

if [ -z "$ESTATE_ID" ]; then
    echo -e "${RED}❌ Failed to create estate${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Estate created:${NC} $ESTATE_ID"
echo ""

# Step 3: ADD TREE
echo -e "${YELLOW}[3/5]${NC} Adding tree to estate..."
TREE_RESPONSE=$(curl -s -X POST "$API_URL/estate/$ESTATE_ID/tree" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "x": 50,
    "y": 60,
    "height": 25
  }')

echo "Response: $TREE_RESPONSE"
echo -e "${GREEN}✓ Tree added${NC}"
echo ""

# Step 4: GET STATS
echo -e "${YELLOW}[4/5]${NC} Getting estate statistics..."
STATS_RESPONSE=$(curl -s -X GET "$API_URL/estate/$ESTATE_ID/stats" \
  -H "Authorization: Bearer $TOKEN")

echo "Response: $STATS_RESPONSE"
echo -e "${GREEN}✓ Stats retrieved${NC}"
echo ""

# Step 5: GET DRONE PLAN
echo -e "${YELLOW}[5/5]${NC} Getting drone plan..."
DRONE_RESPONSE=$(curl -s -X GET "$API_URL/estate/$ESTATE_ID/drone-plan" \
  -H "Authorization: Bearer $TOKEN")

echo "Response: $DRONE_RESPONSE"
echo -e "${GREEN}✓ Drone plan retrieved${NC}"
echo ""

# BONUS: Test without token
echo -e "${YELLOW}[BONUS]${NC} Testing without token (should fail)..."
FAIL_RESPONSE=$(curl -s -X GET "$API_URL/estate/$ESTATE_ID/stats")

echo "Response: $FAIL_RESPONSE"
if echo "$FAIL_RESPONSE" | grep -q "Unauthorized\|Missing authorization"; then
    echo -e "${GREEN}✓ Correctly rejected unauthorized request${NC}"
else
    echo -e "${RED}❌ Should have rejected the request${NC}"
fi

echo ""
echo -e "${GREEN}✅ All tests completed!${NC}"
