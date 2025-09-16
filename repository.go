package main

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Repository struct {
	client    *dynamodb.Client
	tableName string
}

func NewRepository(client *dynamodb.Client, tableName string) *Repository {
	return &Repository{
		client:    client,
		tableName: tableName,
	}
}

// Table Management Operations

func (r *Repository) CreateTable(ctx context.Context) error {
	input := &dynamodb.CreateTableInput{
		TableName: aws.String(r.tableName),
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("pk"), KeyType: types.KeyTypeHash},
			{AttributeName: aws.String("sk"), KeyType: types.KeyTypeRange},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("pk"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("sk"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("status_date"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("placed_id"), AttributeType: types.ScalarAttributeTypeS},
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("inverted-index"),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("sk"), KeyType: types.KeyTypeHash},
					{AttributeName: aws.String("pk"), KeyType: types.KeyTypeRange},
				},
				Projection: &types.Projection{ProjectionType: types.ProjectionTypeAll},
				ProvisionedThroughput: &types.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(5),
					WriteCapacityUnits: aws.Int64(5),
				},
			},
			{
				IndexName: aws.String("placed-index"),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("placed_id"), KeyType: types.KeyTypeHash},
				},
				Projection: &types.Projection{ProjectionType: types.ProjectionTypeAll},
				ProvisionedThroughput: &types.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(5),
					WriteCapacityUnits: aws.Int64(5),
				},
			},
		},
		LocalSecondaryIndexes: []types.LocalSecondaryIndex{
			{
				IndexName: aws.String("status-date-index"),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("pk"), KeyType: types.KeyTypeHash},
					{AttributeName: aws.String("status_date"), KeyType: types.KeyTypeRange},
				},
				Projection: &types.Projection{ProjectionType: types.ProjectionTypeAll},
			},
		},
		BillingMode: types.BillingModeProvisioned,
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
	}

	_, err := r.client.CreateTable(ctx, input)
	return err
}

func (r *Repository) DeleteTable(ctx context.Context) error {
	_, err := r.client.DeleteTable(ctx, &dynamodb.DeleteTableInput{
		TableName: aws.String(r.tableName),
	})
	return err
}

func (r *Repository) EmptyTable(ctx context.Context) error {
	// Scan all items and delete them
	paginator := dynamodb.NewScanPaginator(r.client, &dynamodb.ScanInput{
		TableName: aws.String(r.tableName),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return err
		}

		// Delete items in batches
		for i := 0; i < len(page.Items); i += 25 {
			end := i + 25
			if end > len(page.Items) {
				end = len(page.Items)
			}

			var writeRequests []types.WriteRequest
			for _, item := range page.Items[i:end] {
				writeRequests = append(writeRequests, types.WriteRequest{
					DeleteRequest: &types.DeleteRequest{
						Key: map[string]types.AttributeValue{
							"pk": item["pk"],
							"sk": item["sk"],
						},
					},
				})
			}

			if len(writeRequests) > 0 {
				_, err := r.client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
					RequestItems: map[string][]types.WriteRequest{
						r.tableName: writeRequests,
					},
				})
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// User Operations

func (r *Repository) CreateUser(ctx context.Context, user User) error {
	userMap, err := attributevalue.MarshalMap(user)
	if err != nil {
		return err
	}

	userMap["pk"] = &types.AttributeValueMemberS{Value: fmt.Sprintf("#USER#%s", user.Username)}
	userMap["sk"] = &types.AttributeValueMemberS{Value: "PROFILE"}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      userMap,
	})
	return err
}

func (r *Repository) GetUser(ctx context.Context, username string) (*User, error) {
	result, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		KeyConditionExpression: aws.String("pk = :pk AND sk = :sk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("#USER#%s", username)},
			":sk": &types.AttributeValueMemberS{Value: "PROFILE"},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return nil, err
	}

	if len(result.Items) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	var user User
	if err := attributevalue.UnmarshalMap(result.Items[0], &user); err != nil {
		return nil, err
	}
	user.Username = username
	return &user, nil
}

func (r *Repository) UpdateUser(ctx context.Context, username string, user User) error {
	userMap, err := attributevalue.MarshalMap(user)
	if err != nil {
		return err
	}

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("#USER#%s", username)},
			"sk": &types.AttributeValueMemberS{Value: "PROFILE"},
		},
		AttributeUpdates: make(map[string]types.AttributeValueUpdate),
	}

	for key, value := range userMap {
		if key != "pk" && key != "sk" {
			input.AttributeUpdates[key] = types.AttributeValueUpdate{
				Value:  value,
				Action: types.AttributeActionPut,
			}
		}
	}

	_, err = r.client.UpdateItem(ctx, input)
	return err
}

// Order Operations

func (r *Repository) CreateOrder(ctx context.Context, order *Order) error {
	orderMap, err := attributevalue.MarshalMap(order)
	if err != nil {
		return err
	}

	orderMap["pk"] = &types.AttributeValueMemberS{Value: fmt.Sprintf("#USER#%s", order.UserID)}
	orderMap["sk"] = &types.AttributeValueMemberS{Value: fmt.Sprintf("#ORDER#%s", order.ID)}

	statusDate := fmt.Sprintf("%s#%s", order.Status, order.CreatedAt.Format("2006-01-02"))
	orderMap["status_date"] = &types.AttributeValueMemberS{Value: statusDate}

	if order.Status == OrderStatusPending || order.Status == OrderStatusConfirmed {
		orderMap["placed_id"] = &types.AttributeValueMemberS{Value: string(order.Status)}
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      orderMap,
	})
	return err
}

func (r *Repository) GetOrderByID(ctx context.Context, orderID string) (*Order, error) {
	result, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("inverted-index"),
		KeyConditionExpression: aws.String("sk = :sk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":sk": &types.AttributeValueMemberS{Value: fmt.Sprintf("#ORDER#%s", orderID)},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return nil, err
	}

	if len(result.Items) == 0 {
		return nil, fmt.Errorf("order not found")
	}

	var order Order
	if err := attributevalue.UnmarshalMap(result.Items[0], &order); err != nil {
		return nil, err
	}

	// Extract username from pk
	if pkValue, ok := result.Items[0]["pk"]; ok {
		if pkStr, ok := pkValue.(*types.AttributeValueMemberS); ok {
			order.UserID = pkStr.Value[6:] // Remove "#USER#" prefix
		}
	}
	order.ID = orderID
	return &order, nil
}

func (r *Repository) GetOrdersByUserID(ctx context.Context, userID string) ([]*Order, error) {
	result, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		KeyConditionExpression: aws.String("pk = :pk AND begins_with(sk, :sk_prefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":        &types.AttributeValueMemberS{Value: fmt.Sprintf("#USER#%s", userID)},
			":sk_prefix": &types.AttributeValueMemberS{Value: "#ORDER#"},
		},
		ScanIndexForward: aws.Bool(false),
	})
	if err != nil {
		return nil, err
	}

	var orders []*Order
	for _, item := range result.Items {
		var order Order
		if err := attributevalue.UnmarshalMap(item, &order); err != nil {
			continue
		}
		order.UserID = userID
		if skValue, ok := item["sk"]; ok {
			if skStr, ok := skValue.(*types.AttributeValueMemberS); ok {
				order.ID = skStr.Value[7:] // Remove "#ORDER#" prefix
			}
		}
		orders = append(orders, &order)
	}

	return orders, nil
}

func (r *Repository) UpdateOrderStatus(ctx context.Context, orderID string, status OrderStatus) error {
	order, err := r.GetOrderByID(ctx, orderID)
	if err != nil {
		return err
	}

	statusDate := fmt.Sprintf("%s#%s", status, time.Now().Format("2006-01-02"))

	updateExpression := "SET #status = :status, #status_date = :status_date, #updated_at = :updated_at"
	expressionAttributeNames := map[string]string{
		"#status":      "status",
		"#status_date": "status_date",
		"#updated_at":  "updated_at",
	}
	expressionAttributeValues := map[string]types.AttributeValue{
		":status":      &types.AttributeValueMemberS{Value: string(status)},
		":status_date": &types.AttributeValueMemberS{Value: statusDate},
		":updated_at":  &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
	}

	if status == OrderStatusPending || status == OrderStatusConfirmed {
		updateExpression += ", #placed_id = :placed_id"
		expressionAttributeNames["#placed_id"] = "placed_id"
		expressionAttributeValues[":placed_id"] = &types.AttributeValueMemberS{Value: string(status)}
	} else {
		updateExpression += " REMOVE #placed_id"
		expressionAttributeNames["#placed_id"] = "placed_id"
	}

	_, err = r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("#USER#%s", order.UserID)},
			"sk": &types.AttributeValueMemberS{Value: fmt.Sprintf("#ORDER#%s", orderID)},
		},
		UpdateExpression:          aws.String(updateExpression),
		ExpressionAttributeNames:  expressionAttributeNames,
		ExpressionAttributeValues: expressionAttributeValues,
	})

	return err
}

func (r *Repository) GetPendingOrders(ctx context.Context) ([]*Order, error) {
	result, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("placed-index"),
		KeyConditionExpression: aws.String("placed_id = :placed_id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":placed_id": &types.AttributeValueMemberS{Value: string(OrderStatusPending)},
		},
	})
	if err != nil {
		return nil, err
	}

	var orders []*Order
	for _, item := range result.Items {
		var order Order
		if err := attributevalue.UnmarshalMap(item, &order); err != nil {
			continue
		}
		orders = append(orders, &order)
	}

	return orders, nil
}

// Order Item Operations

func (r *Repository) CreateOrderItem(ctx context.Context, orderID string, item *OrderItem) error {
	itemMap, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err
	}

	itemMap["pk"] = &types.AttributeValueMemberS{Value: fmt.Sprintf("#ORDER#%s", orderID)}
	itemMap["sk"] = &types.AttributeValueMemberS{Value: fmt.Sprintf("#ITEM#%s", item.ItemID)}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      itemMap,
	})
	return err
}

func (r *Repository) GetOrderItems(ctx context.Context, orderID string) ([]OrderItem, error) {
	result, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		KeyConditionExpression: aws.String("pk = :pk AND begins_with(sk, :sk_prefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":        &types.AttributeValueMemberS{Value: fmt.Sprintf("#ORDER#%s", orderID)},
			":sk_prefix": &types.AttributeValueMemberS{Value: "#ITEM#"},
		},
	})
	if err != nil {
		return nil, err
	}

	var items []OrderItem
	for _, item := range result.Items {
		var orderItem OrderItem
		if err := attributevalue.UnmarshalMap(item, &orderItem); err != nil {
			continue
		}
		items = append(items, orderItem)
	}

	return items, nil
}
