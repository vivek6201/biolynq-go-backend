package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/vivek6201/biolynq/internal/bootstrap/server"
	"github.com/vivek6201/biolynq/internal/config"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file:", err)
	}
	cfg := config.LoadConfig()
	server.StartServer(cfg)
}
