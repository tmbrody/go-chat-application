package main

import (
	"context"
	"go-chat-application/config"
	"go-chat-application/internal/database"
	"go-chat-application/routes"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Load environment variables from .env file
	godotenv.Load()

	// Get PORT environment variable
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable is not set")
	}

	// Get DATABASE_URL environment variable
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}
	// Extract database name from the database URL
	dbName, err := database.GetDatabaseNAmeFromURL(dbURL)
	if err != nil {
		log.Fatal("Could not parse the database url")
	}

	// Get JWT_SECRET environment variable
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// Ensure the context is cancelled to avoid leaking resources
	defer cancel()
	// Connect to the MongoDB database
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dbURL))
	// Check if connection to the MongoDB database was successful
	if err != nil {
		log.Fatal("Could not open a connection to the MongoDB database")
	}

	// Create a MongoDB client
	mongoClient := &database.MongoDBClient{Client: client, DBName: dbName}

	// Set the database and JWT secret in the API configuration
	config.ApiCfg.DB = mongoClient
	config.ApiCfg.JwtSecret = jwtSecret

	// Create new routers
	r := chi.NewRouter()
	r_api := chi.NewRouter()

	// Setup CORS for the main router
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Setup CORS for the API router
	r_api.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Mount the API router on the main router
	r.Mount("/api", r_api)

	// Setup the routes for the main and API routers
	routes.SetupRoutes(r)
	routes.SetupApiRoutes(r_api)

	// Create a new HTTP server
	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Start the HTTP server
	err = server.ListenAndServe()
	// Check if the server started successfully
	if err != nil {
		log.Fatal(err)
	}
}
