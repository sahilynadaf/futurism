package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gorilla/mux"
)

// Port represents the structure of a port record
type Port struct {
	Name        string    `json:"name"`
	City        string    `json:"city"`
	Country     string    `json:"country"`
	Alias       []string  `json:"alias"`
	Regions     []string  `json:"regions"`
	Coordinates []float64 `json:"coordinates"`
	Province    string    `json:"province"`
	Timezone    string    `json:"timezone"`
	Unlocs      []string  `json:"unlocs"`
	Code        string    `json:"code"`
}

// Database is an in-memory store for ports
type Database struct {
	ports map[string]Port
	mu    sync.RWMutex
}

// NewDatabase initializes an empty database
func NewDatabase() *Database {
	return &Database{ports: make(map[string]Port)}
}

// Upsert inserts or updates a port record
func (db *Database) Upsert(key string, port Port) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.ports[key] = port
}

// GetPort retrieves a port record by key
func (db *Database) GetPort(key string) (Port, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	port, exists := db.ports[key]
	return port, exists
}

// LoadPorts reads and processes the ports.json file
func LoadPorts(db *Database, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var data map[string]Port

	if err := decoder.Decode(&data); err != nil {
		return err
	}

	for key, port := range data {
		db.Upsert(key, port)
	}

	return nil
}

// Handler functions
func getPortHandler(db *Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		key := vars["id"]
		if port, exists := db.GetPort(key); exists {
			json.NewEncoder(w).Encode(port)
		} else {
			http.Error(w, "Port not found", http.StatusNotFound)
		}
	}
}

func main() {
	db := NewDatabase()
	filename := "ports.json"

	if err := LoadPorts(db, filename); err != nil {
		log.Fatal("Error loading ports:", err)
	}

	log.Println("Successfully loaded ports into memory.")

	r := mux.NewRouter()
	r.HandleFunc("/ports/{id}", getPortHandler(db)).Methods("GET")

	httpServer := &http.Server{Addr: ":8080", Handler: r}
	go func() {
		log.Println("Starting server on :8080")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not start server: %v", err)
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	log.Println("Shutting down gracefully...")
	os.Exit(0)
}
