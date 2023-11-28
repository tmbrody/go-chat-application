package config

import (
	"go-chat-application/internal/database"
)

type ApiConfig struct {
	DB        *database.MongoDBClient
	JwtSecret string
}

var ApiCfg ApiConfig

type ContextKey string

var DBContextKey ContextKey = "db"
