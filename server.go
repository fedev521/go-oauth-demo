package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"example.com/go-oauth/mw"
	"github.com/joho/godotenv"
)

func main() {
	// load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading the .env file: %v", err)
	}

	// create router
	router := mw.NewRouter()

	// create server with configuration
	server := &http.Server{
		Addr:         "0.0.0.0" + ":" + os.Getenv("PORT"),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      router,
	}
	crtFile := "server.crt"
	keyFile := "server.key"

	// start server
	log.Fatal(server.ListenAndServeTLS(crtFile, keyFile))
}
