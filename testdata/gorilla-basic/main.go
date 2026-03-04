package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// CreateUserRequest is the request body for creating a user.
type CreateUserRequest struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required"`
	Age   int    `json:"age,omitempty"`
}

// UserResponse is the response body for user endpoints.
type UserResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age,omitempty"`
}

// ErrorResponse is returned for all error cases.
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Details string `json:"details,omitempty"`
}

// ItemResponse is the response for item endpoints.
type ItemResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func main() {
	r := mux.NewRouter()

	// Middleware on root router
	r.Use(loggingMiddleware)

	// Direct route registrations with .Methods()
	r.HandleFunc("/users", ListUsers).Methods("GET")
	r.HandleFunc("/users", CreateUser).Methods("POST")
	r.HandleFunc("/users/{id}", GetUser).Methods("GET")
	r.HandleFunc("/users/{id}", DeleteUser).Methods("DELETE")

	// Route without .Methods() → ANY
	r.HandleFunc("/health", HealthCheck)

	// Subrouter with PathPrefix
	api := r.PathPrefix("/api/v1").Subrouter()
	api.Use(authMiddleware)
	api.HandleFunc("/items", ListItems).Methods("GET")
	api.HandleFunc("/items/{id:[0-9]+}", GetItem).Methods("GET")

	// Nested subrouter
	admin := r.PathPrefix("/admin").Subrouter()
	admin.Use(authMiddleware)
	admin.HandleFunc("/dashboard", GetDashboard).Methods("GET")

	_ = http.ListenAndServe(":3000", r)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" || !strings.HasPrefix(token, "Bearer ") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(ErrorResponse{
				Error: "unauthorized",
				Code:  http.StatusUnauthorized,
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ListUsers returns a paginated list of users.
func ListUsers(w http.ResponseWriter, r *http.Request) {
	page := r.URL.Query().Get("page")
	if page == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: "page parameter is required",
			Code:  http.StatusBadRequest,
		})
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err == nil {
			limit = parsed
		}
	}
	_ = limit

	pageNum, _ := strconv.Atoi(page)
	users := []UserResponse{
		{ID: fmt.Sprintf("user-%d", pageNum), Name: "Alice", Email: "alice@example.com", Age: 30},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(users)
}

// GetUser returns a single user by ID.
func GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: "missing user id",
			Code:  http.StatusBadRequest,
		})
		return
	}

	user := UserResponse{
		ID:    id,
		Name:  "Alice",
		Email: "alice@example.com",
		Age:   30,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(user)
}

// CreateUser creates a new user from the JSON request body.
func CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "invalid request body",
			Code:    http.StatusBadRequest,
			Details: err.Error(),
		})
		return
	}

	user := UserResponse{
		ID:    "new-user-id",
		Name:  req.Name,
		Email: req.Email,
		Age:   req.Age,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(user)
}

// DeleteUser deletes a user by ID.
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: "missing user id",
			Code:  http.StatusBadRequest,
		})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// HealthCheck returns service health status.
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// ListItems returns all items.
func ListItems(w http.ResponseWriter, r *http.Request) {
	items := []ItemResponse{
		{ID: "1", Name: "Widget"},
		{ID: "2", Name: "Gadget"},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(items)
}

// GetItem returns a single item by numeric ID.
func GetItem(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	intID, err := strconv.Atoi(id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: "invalid item id",
			Code:  http.StatusBadRequest,
		})
		return
	}

	item := ItemResponse{
		ID:   strconv.Itoa(intID),
		Name: "Widget",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(item)
}

// GetDashboard returns admin dashboard data.
func GetDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"page": "dashboard"})
}
