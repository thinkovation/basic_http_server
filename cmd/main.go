package main

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

var Logger *slog.Logger

func main() {
	Logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	godotenv.Load()

	server := NewServer(os.Getenv("HTTP_PORT"))

	Logger.Info("Starting Server on port " + os.Getenv("HTTP_PORT"))

	go func() {
		// This starts the HTTP server
		err := server.ListenAndServe()

		if err != nil {
			Logger.Error("Exiting: " + err.Error())
		}
	}()

	//wait shutdown
	server.WaitShutdown()

	Logger.Info("Service Exiting")
}
