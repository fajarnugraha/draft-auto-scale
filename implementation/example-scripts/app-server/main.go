package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

// --- Global Configuration ---
var (
	// sessions holds the active user sessions using a concurrent map.
	sessions sync.Map

	// loadConfig holds the CPU and memory load simulation settings.
	loadConfig struct {
		CPUIterations int
		MemMB         int
	}

	// memoryStore holds a pre-allocated slice of bytes to simulate memory usage
	// without causing constant garbage collection.
	memoryStore []byte
)

// session represents a user session.
type session struct {
	username  string
	expiresAt time.Time
}

// --- Utility Functions ---

// simulateLoad generates CPU and memory load.
func simulateLoad(cpuIterations int, memMB int) {
	// 1. CPU Load: Perform some calculations.
	for i := 0; i < cpuIterations; i++ {
		_ = math.Sqrt(float64(i))
	}

	// 2. Memory Load: Access the pre-allocated slice to simulate usage.
	if len(memoryStore) > 0 {
		// Iterate through the slice to ensure it's paged into RAM.
		for i := 0; i < len(memoryStore); i += 1024 { // Step by 1KB to be efficient
			memoryStore[i] = byte(i % 256)
		}
	}
}

// isAuthenticated checks if a session token from the request is valid.
func isAuthenticated(ctx *fasthttp.RequestCtx) bool {
	authHeader := ctx.Request.Header.Peek("Authorization")
	if len(authHeader) == 0 {
		return false
	}

	// The token is expected to be in the format "Bearer <token>"
	tokenBytes := bytes.TrimPrefix(authHeader, []byte("Bearer "))

	s, ok := sessions.Load(string(tokenBytes))
	if !ok {
		return false
	}

	// Type assertion
	sess, ok := s.(session)
	if !ok {
		// This case should not happen if we only store `session` types.
		return false
	}

	return time.Now().Before(sess.expiresAt)
}

// --- HTTP Handlers ---

// loginHandler handles user login, creates a session, and simulates load.
func loginHandler(ctx *fasthttp.RequestCtx) {
	if !ctx.IsPost() {
		ctx.Error("Invalid request method", fasthttp.StatusMethodNotAllowed)
		return
	}

	// Simulate work before processing the request.
	simulateLoad(loadConfig.CPUIterations, loadConfig.MemMB)

	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.Unmarshal(ctx.PostBody(), &creds); err != nil {
		ctx.Error("Invalid request body", fasthttp.StatusBadRequest)
		return
	}

	if creds.Username == "" {
		ctx.Error("Username is required", fasthttp.StatusBadRequest)
		return
	}

	sessionToken := uuid.NewString()
	expiresAt := time.Now().Add(120 * time.Second)

	// Store the new session in the sync.Map.
	log.Printf("User '%s' logged in. Session token: %s", creds.Username, sessionToken)
	sessions.Store(sessionToken, session{
		username:  creds.Username,
		expiresAt: expiresAt,
	})

	ctx.SetContentType("application/json")
	json.NewEncoder(ctx).Encode(map[string]string{
		"token": sessionToken,
	})
	log.Printf("User '%s' logged in. Session token: %s", creds.Username, sessionToken)
}

// browseHandler simulates browsing products and generates load.
func browseHandler(ctx *fasthttp.RequestCtx) {
	if !isAuthenticated(ctx) {
		ctx.Error("Unauthorized", fasthttp.StatusUnauthorized)
		return
	}

	// Simulate work.
	simulateLoad(loadConfig.CPUIterations, loadConfig.MemMB)
	// time.Sleep(100 * time.Millisecond) // Additional latency

	ctx.SetContentType("application/json")
	json.NewEncoder(ctx).Encode(map[string]interface{}{
		"status": "success",
		"data": []string{
			"Product A",
			"Product B",
			"Product C",
		},
	})
	log.Println("Browse endpoint accessed.")
}

// submitHandler simulates submitting data and generates load.
func submitHandler(ctx *fasthttp.RequestCtx) {
	if !isAuthenticated(ctx) {
		ctx.Error("Unauthorized", fasthttp.StatusUnauthorized)
		return
	}

	if !ctx.IsPost() {
		ctx.Error("Invalid request method", fasthttp.StatusMethodNotAllowed)
		return
	}

	// Simulate work.
	simulateLoad(loadConfig.CPUIterations, loadConfig.MemMB)

	var data map[string]interface{}
	if err := json.Unmarshal(ctx.PostBody(), &data); err != nil {
		ctx.Error("Invalid request body", fasthttp.StatusBadRequest)
		return
	}

	// time.Sleep(200 * time.Millisecond) // Additional latency

	ctx.SetContentType("application/json")
	json.NewEncoder(ctx).Encode(map[string]string{
		"status":  "success",
		"message": "Data submitted successfully",
	})
	log.Printf("Submit endpoint accessed with data: %v", data)
}

// metricsHandler exposes the number of active sessions.
func metricsHandler(ctx *fasthttp.RequestCtx) {
	activeSessions := 0
	// Iterate over the sync.Map and clean up expired sessions.
	sessions.Range(func(key, value interface{}) bool {
		// Type assertion
		sess, ok := value.(session)
		if !ok {
			return true // Continue to next item
		}

		if time.Now().After(sess.expiresAt) {
			// Delete expired session.
			sessions.Delete(key)
		} else {
			// Count active sessions.
			activeSessions++
		}
		return true // Continue to next item
	})

	// Prometheus format
	fmt.Fprintf(ctx, "# HELP concurrent_connections The number of active user sessions.\n")
	fmt.Fprintf(ctx, "# TYPE concurrent_connections gauge\n")
	fmt.Fprintf(ctx, "concurrent_connections %d\n", activeSessions)
}

// --- Main Function ---

func main() {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Enable the randomness pool for UUID generation.
	uuid.EnableRandPool()

	// Load configuration from environment variables with defaults.
	var err error
	loadConfig.CPUIterations, err = strconv.Atoi(os.Getenv("LOAD_CPU_ITERATIONS"))
	if err != nil {
		loadConfig.CPUIterations = 0 // Default to 0 if not set
	}

	loadConfig.MemMB, err = strconv.Atoi(os.Getenv("LOAD_MEM_MB"))
	if err != nil {
		loadConfig.MemMB = 0 // Default to 0 if not set
	}

	// Pre-allocate memory store if configured
	if loadConfig.MemMB > 0 {
		log.Printf("Pre-allocating %d MB of memory...", loadConfig.MemMB)
		memoryStore = make([]byte, loadConfig.MemMB*1024*1024)
	}


	// The router is responsible for matching incoming requests to their corresponding handler.
	router := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/login":
			loginHandler(ctx)
		case "/browse":
			browseHandler(ctx)
		case "/submit":
			submitHandler(ctx)
		case "/metrics":
			metricsHandler(ctx)
		default:
			ctx.Error("Unsupported path", fasthttp.StatusNotFound)
		}
	}

	log.Printf("Server starting on port 8080...")
	log.Printf("Load simulation settings: CPU Iterations=%d, Memory MB=%d", loadConfig.CPUIterations, loadConfig.MemMB)

	if err := fasthttp.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}