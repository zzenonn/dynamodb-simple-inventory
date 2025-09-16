# Simple DynamoDB Inventory - Learning Project

This is a simplified version focused on DynamoDB data modeling with a basic REST API. Perfect for learning DynamoDB fundamentals without complex architecture patterns.

## What You'll Learn

- **Single Table Design**: How to store multiple entity types in one DynamoDB table
- **Composite Keys**: Using partition key (pk) and sort key (sk) for data organization
- **Access Patterns**: How to design keys to support your query needs
- **Global Secondary Indexes (GSI)**: Creating alternate access patterns
- **Local Secondary Indexes (LSI)**: Additional sort options within partition

## Data Model Overview

We use a single DynamoDB table to store:
- **Users**: Customer profiles with addresses
- **Orders**: Order headers with status tracking
- **Order Items**: Individual items within orders

### Key Design Patterns

#### 1. Composite Keys
```
Users:     pk="#USER#<username>"    sk="PROFILE"
Orders:    pk="#USER#<username>"    sk="#ORDER#<orderid>"
Items:     pk="#ORDER#<orderid>"    sk="#ITEM#<itemid>"
```

#### 2. Access Patterns Supported
- Get user profile by username
- Get all orders for a user
- Get order by order ID (using GSI)
- Get order items for an order
- Get orders by status and date range
- Get pending orders across all users

#### 3. Indexes Used
- **Main Table**: pk + sk
- **inverted-index (GSI)**: sk + pk (find orders by order ID)
- **status-date-index (LSI)**: pk + status_date (orders by status/date)
- **placed-index (GSI)**: placed_id (sparse index for pending orders)

## Environment Setup

Set these environment variables:
```bash
export AWS_REGION=us-east-1
export DYNAMODB_TABLE_NAME=simple-inventory
export AWS_PROFILE=your-profile  # or use AWS credentials
```

## Table Management Commands

```bash
# Create the DynamoDB table with indexes
go run . -create-table

# Empty all data from the table (keeps table structure)
go run . -empty-table

# Delete the entire table
go run . -delete-table
```

## Running the API Server

```bash
# Start the server (default port 8080)
go run .

# Start on different port
go run . -port=3000
```

## API Endpoints

```
POST   /users              - Create user
GET    /users/{username}   - Get user profile
PUT    /users/{username}   - Update user profile

POST   /orders             - Create order
GET    /orders/{orderid}   - Get order by ID
GET    /users/{username}/orders - Get user's orders
PUT    /orders/{orderid}/status - Update order status

POST   /orders/{orderid}/items - Add item to order
GET    /orders/{orderid}/items - Get order items

GET    /orders/pending     - Get all pending orders
```

## Quick Demo

1. **Setup the table:**
   ```bash
   go run . -create-table
   ```

2. **Start the server:**
   ```bash
   go run .
   ```

3. **Run the examples:**
   ```bash
   ./examples.sh
   ```

## Example Usage

```bash
# Create a user
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{
    "username":"john",
    "email":"john@example.com",
    "full_name":"John Doe",
    "addresses": {
      "home": {
        "street": "123 Main St",
        "state": "CA",
        "country": "USA"
      }
    }
  }'

# Create an order
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"user_id":"john","address_key":"home"}'

# Get user orders
curl http://localhost:8080/users/john/orders
```

## DynamoDB Learning Points

### Single Table Design Benefits
- **Cost Effective**: One table instead of multiple
- **Performance**: Related data co-located
- **Atomic Operations**: Transactions within partition

### Key Patterns Demonstrated
1. **Hierarchical Data**: User → Orders → Items
2. **Sparse Indexes**: Only pending orders in placed-index
3. **Composite Sort Keys**: status#date for time-based queries
4. **Inverted Index**: Query by order ID across all users

### Access Pattern Examples
- **Get User Profile**: Query pk="#USER#john" AND sk="PROFILE"
- **Get User Orders**: Query pk="#USER#john" AND begins_with(sk, "#ORDER#")
- **Get Order by ID**: Query inverted-index where sk="#ORDER#uuid"
- **Get Pending Orders**: Query placed-index where placed_id="pending"

## Files Overview

- `main.go` - Entry point with CLI commands and server setup
- `models.go` - Domain models (User, Order, OrderItem)
- `repository.go` - DynamoDB operations and table management
- `handlers.go` - HTTP API handlers
- `examples.sh` - Demo script showing all operations
