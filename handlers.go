package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type API struct {
	repo *Repository
}

func NewAPI(repo *Repository) *API {
	return &API{repo: repo}
}

// User handlers

func (api *API) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if user.Username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	if err := api.repo.CreateUser(r.Context(), user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (api *API) GetUser(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	
	user, err := api.repo.GetUser(r.Context(), username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (api *API) UpdateUser(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := api.repo.UpdateUser(r.Context(), username, user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user.Username = username
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// Order handlers

func (api *API) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID     string `json:"user_id"`
		AddressKey string `json:"address_key"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.UserID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	order := &Order{
		ID:         uuid.New().String(),
		UserID:     req.UserID,
		Status:     OrderStatusPending,
		AddressKey: req.AddressKey,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := api.repo.CreateOrder(r.Context(), order); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func (api *API) GetOrder(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "orderid")
	
	order, err := api.repo.GetOrderByID(r.Context(), orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func (api *API) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	
	orders, err := api.repo.GetOrdersByUserID(r.Context(), username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func (api *API) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "orderid")
	
	var req struct {
		Status OrderStatus `json:"status"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := api.repo.UpdateOrderStatus(r.Context(), orderID, req.Status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (api *API) GetPendingOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := api.repo.GetPendingOrders(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

// Order Item handlers

func (api *API) CreateOrderItem(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "orderid")
	
	var item OrderItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if item.ItemID == "" {
		item.ItemID = uuid.New().String()
	}
	item.OrderID = orderID

	if err := api.repo.CreateOrderItem(r.Context(), orderID, &item); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

func (api *API) GetOrderItems(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "orderid")
	
	items, err := api.repo.GetOrderItems(r.Context(), orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}
