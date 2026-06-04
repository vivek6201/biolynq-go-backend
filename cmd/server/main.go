package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/vivek6201/biolynq/internal/bootstrap/server"
	"github.com/vivek6201/biolynq/internal/config"
)

func main() {
	cfg := config.LoadConfig()
	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file:", err)
	}
	server.StartServer(cfg)
}
