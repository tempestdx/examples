package main

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/tempestdx/examples/deps/dashboards/server/models"
)

var (
	dashboards = []*models.Dashboard{}
	mu         sync.Mutex

	logger = slog.New(slog.NewJSONHandler(os.Stderr, nil))
)

// Generate a unique alphanumeric ExternalID.
func generateExternalID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 8
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Int()%len(charset)]
	}

	return string(b)
}

func lookupDashboard(externalID string) *models.Dashboard {
	for _, dashboard := range dashboards {
		if dashboard.ID == externalID {
			return dashboard
		}
	}
	return nil
}

func createDashboard(w http.ResponseWriter, r *http.Request) {
	var dashboard models.Dashboard
	err := json.NewDecoder(r.Body).Decode(&dashboard)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if dashboard.Project == "" {
		http.Error(w, "Project is required", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()
	dashboard.ID = generateExternalID()
	dashboards = append(dashboards, &dashboard)

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(dashboard); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func getDashboard(w http.ResponseWriter, r *http.Request) {
	externalID := r.URL.Query().Get("id")
	mu.Lock()
	defer mu.Unlock()

	dashboard := lookupDashboard(externalID)
	if dashboard == nil {
		http.Error(w, "Dashboard not found", http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(dashboard); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func updateDashboard(w http.ResponseWriter, r *http.Request) {
	externalID := r.URL.Query().Get("id")

	var updatedDashboard models.Dashboard
	err := json.NewDecoder(r.Body).Decode(&updatedDashboard)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()
	dashboard := lookupDashboard(externalID)
	if dashboard == nil {
		http.Error(w, "Dashboard not found", http.StatusNotFound)
		return
	}

	dashboard.Name = updatedDashboard.Name
	dashboard.Description = updatedDashboard.Description

	if err := json.NewEncoder(w).Encode(dashboard); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func deleteDashboard(w http.ResponseWriter, r *http.Request) {
	externalID := r.URL.Query().Get("id")
	mu.Lock()
	defer mu.Unlock()
	dashboard := lookupDashboard(externalID)
	if dashboard == nil {
		http.Error(w, "Dashboard not found", http.StatusNotFound)
		return
	}

	for i, d := range dashboards {
		if d.ID == externalID {
			dashboards = append(dashboards[:i], dashboards[i+1:]...)
			break
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func listDashboards(w http.ResponseWriter, r *http.Request) {
	const maxres = 2

	next := r.URL.Query().Get("next")
	var nextInt int
	if next != "" {
		var err error
		nextInt, err = strconv.Atoi(next)
		if err != nil {
			http.Error(w, "Invalid next value", http.StatusBadRequest)
			return
		}
	}

	mu.Lock()
	defer mu.Unlock()

	var dashboardList models.DashboardList
	for i := nextInt; i < nextInt+maxres; i++ {
		if i >= len(dashboards) {
			break
		}
		dashboardList.Dashboards = append(dashboardList.Dashboards, *dashboards[i])
	}
	if nextInt+maxres < len(dashboards) {
		dashboardList.Next = nextInt + maxres
	}

	if err := json.NewEncoder(w).Encode(dashboardList); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Request received", "path", r.URL.Path, "params", r.URL.Query())
		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/dashboard/create", createDashboard)
	mux.HandleFunc("/dashboard/get", getDashboard)
	mux.HandleFunc("/dashboard/update", updateDashboard)
	mux.HandleFunc("/dashboard/delete", deleteDashboard)
	mux.HandleFunc("/dashboard/list", listDashboards)
	mux.HandleFunc("/healthz", healthz)

	loggedMux := loggingMiddleware(mux)

	fmt.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", loggedMux))
}
