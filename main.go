package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//go:embed static/config.json static/templates/*.html
var staticFiles embed.FS

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	// Generate a unique server ID
	serverID := uuid.New()
	fmt.Printf("Starting server with ID: %s\n", serverID)

	// Initialize logger
	logger := logrus.New()
	logger.Info("Starting application")

	// Initialize database connections (dummy for transitive deps)
	ctx := context.Background()
	mongoClient, _ := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	defer mongoClient.Disconnect(ctx)

	db, _ := gorm.Open(postgres.Open("host=localhost"), &gorm.Config{})
	_ = db

	// Initialize Gin router
	router := gin.Default()
	router.GET("/api/status", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Set up HTTP routes
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/health", handleHealth)

	// Start server
	port := ":8080"
	fmt.Printf("Server listening on %s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to go-test server!\n")
	fmt.Fprintf(w, "Time: %s\n", time.Now().Format(time.RFC3339))
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	// Echo server
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket read error:", err)
			break
		}

		log.Printf("Received: %s", message)

		err = conn.WriteMessage(messageType, message)
		if err != nil {
			log.Println("WebSocket write error:", err)
			break
		}
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
}
