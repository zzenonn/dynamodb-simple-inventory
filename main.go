package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	var (
		createTable = flag.Bool("create-table", false, "Create DynamoDB table")
		deleteTable = flag.Bool("delete-table", false, "Delete DynamoDB table")
		emptyTable  = flag.Bool("empty-table", false, "Empty DynamoDB table")
		port        = flag.String("port", "8080", "Server port")
	)
	flag.Parse()

	// Get configuration from environment
	tableName := os.Getenv("DYNAMODB_TABLE_NAME")
	if tableName == "" {
		tableName = "simple-inventory"
	}

	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}

	// Initialize AWS config
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
	)
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	// Create DynamoDB client
	client := dynamodb.NewFromConfig(cfg)
	repo := NewRepository(client, tableName)

	ctx := context.Background()

	// Handle CLI commands
	if *createTable {
		fmt.Printf("Creating table '%s'...\n", tableName)
		if err := repo.CreateTable(ctx); err != nil {
			log.Fatalf("Failed to create table: %v", err)
		}
		fmt.Println("Table created successfully!")
		return
	}

	if *deleteTable {
		fmt.Printf("Deleting table '%s'...\n", tableName)
		if err := repo.DeleteTable(ctx); err != nil {
			log.Fatalf("Failed to delete table: %v", err)
		}
		fmt.Println("Table deleted successfully!")
		return
	}

	if *emptyTable {
		fmt.Printf("Emptying table '%s'...\n", tableName)
		if err := repo.EmptyTable(ctx); err != nil {
			log.Fatalf("Failed to empty table: %v", err)
		}
		fmt.Println("Table emptied successfully!")
		return
	}

	// Start API server
	api := NewAPI(repo)
	r := setupRoutes(api)

	fmt.Printf("Starting server on port %s...\n", *port)
	fmt.Printf("Table: %s\n", tableName)
	fmt.Printf("Region: %s\n", region)
	fmt.Println("\nAPI Endpoints:")
	fmt.Println("POST   /users              - Create user")
	fmt.Println("GET    /users/{username}   - Get user profile")
	fmt.Println("PUT    /users/{username}   - Update user profile")
	fmt.Println("POST   /orders             - Create order")
	fmt.Println("GET    /orders/{orderid}   - Get order by ID")
	fmt.Println("GET    /users/{username}/orders - Get user's orders")
	fmt.Println("PUT    /orders/{orderid}/status - Update order status")
	fmt.Println("POST   /orders/{orderid}/items - Add item to order")
	fmt.Println("GET    /orders/{orderid}/items - Get order items")
	fmt.Println("GET    /orders/pending     - Get all pending orders")

	log.Fatal(http.ListenAndServe(":"+*port, r))
}

func setupRoutes(api *API) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.SetHeader("Content-Type", "application/json"))

	// User routes
	r.Post("/users", api.CreateUser)
	r.Get("/users/{username}", api.GetUser)
	r.Put("/users/{username}", api.UpdateUser)
	r.Get("/users/{username}/orders", api.GetUserOrders)

	// Order routes
	r.Post("/orders", api.CreateOrder)
	r.Get("/orders/{orderid}", api.GetOrder)
	r.Put("/orders/{orderid}/status", api.UpdateOrderStatus)
	r.Get("/orders/pending", api.GetPendingOrders)

	// Order item routes
	r.Post("/orders/{orderid}/items", api.CreateOrderItem)
	r.Get("/orders/{orderid}/items", api.GetOrderItems)

	return r
}
