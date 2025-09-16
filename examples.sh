#!/bin/bash

# Simple DynamoDB Inventory - Example Usage
# Make sure the server is running: go run .

BASE_URL="http://localhost:8080"

echo "=== DynamoDB Learning Examples ==="
echo

echo "1. Creating a user..."
curl -X POST $BASE_URL/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john",
    "full_name": "John Doe",
    "email": "john@example.com",
    "addresses": {
      "home": {
        "street": "123 Main St",
        "state": "CA",
        "country": "USA"
      },
      "work": {
        "street": "456 Office Blvd",
        "state": "CA", 
        "country": "USA"
      }
    }
  }'
echo -e "\n"

echo "2. Getting user profile..."
curl $BASE_URL/users/john
echo -e "\n"

echo "3. Creating an order..."
ORDER_RESPONSE=$(curl -s -X POST $BASE_URL/orders \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "john",
    "address_key": "home"
  }')
echo $ORDER_RESPONSE

# Extract order ID for next steps
ORDER_ID=$(echo $ORDER_RESPONSE | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
echo "Order ID: $ORDER_ID"
echo

echo "4. Adding items to the order..."
curl -X POST $BASE_URL/orders/$ORDER_ID/items \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Laptop",
    "description": "Gaming laptop",
    "price": 1299.99,
    "quantity": 1
  }'
echo -e "\n"

curl -X POST $BASE_URL/orders/$ORDER_ID/items \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Mouse",
    "description": "Wireless mouse",
    "price": 29.99,
    "quantity": 2
  }'
echo -e "\n"

echo "5. Getting order details..."
curl $BASE_URL/orders/$ORDER_ID
echo -e "\n"

echo "6. Getting order items..."
curl $BASE_URL/orders/$ORDER_ID/items
echo -e "\n"

echo "7. Getting user's orders..."
curl $BASE_URL/users/john/orders
echo -e "\n"

echo "8. Updating order status..."
curl -X PUT $BASE_URL/orders/$ORDER_ID/status \
  -H "Content-Type: application/json" \
  -d '{"status": "confirmed"}'
echo -e "\n"

echo "9. Getting all pending orders..."
curl $BASE_URL/orders/pending
echo -e "\n"

echo "=== DynamoDB Data Model Explanation ==="
echo
echo "This demonstrates DynamoDB single-table design:"
echo "- Users stored as: pk='#USER#john', sk='PROFILE'"
echo "- Orders stored as: pk='#USER#john', sk='#ORDER#<uuid>'"
echo "- Items stored as: pk='#ORDER#<uuid>', sk='#ITEM#<uuid>'"
echo
echo "Access patterns supported:"
echo "- Get user by username (main table)"
echo "- Get user's orders (main table query)"
echo "- Get order by ID (inverted-index GSI)"
echo "- Get order items (main table query)"
echo "- Get pending orders (placed-index GSI)"
echo "- Query orders by status/date (status-date-index LSI)"
