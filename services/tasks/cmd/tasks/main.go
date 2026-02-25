package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pz1.2/services/tasks/internal/client/authclient"
	taskshttp "pz1.2/services/tasks/internal/http"
	"pz1.2/services/tasks/internal/service"
	"pz1.2/shared/middleware"
)

func main() {
	port := os.Getenv("TASKS_PORT")
	if port == "" {
		port = "8082"
	}

	authMode := os.Getenv("AUTH_MODE")
	if authMode == "" {
		authMode = "http"
	}

	var authVerifier authclient.AuthVerifier

	switch authMode {
	case "grpc":
		grpcAddr := os.Getenv("AUTH_GRPC_ADDR")
		if grpcAddr == "" {
			grpcAddr = "localhost:50051"
		}
		log.Printf("Using gRPC auth client, connecting to %s", grpcAddr)
		client, err := authclient.NewGRPCClient(grpcAddr, 2*time.Second)
		if err != nil {
			log.Fatalf("Failed to create gRPC auth client: %v", err)
		}
		authVerifier = client
		defer client.Close()
	default:
		authBaseURL := os.Getenv("AUTH_BASE_URL")
		if authBaseURL == "" {
			authBaseURL = "http://localhost:8081"
		}
		log.Printf("Using HTTP auth client, connecting to %s", authBaseURL)
		authVerifier = authclient.NewHTTPClient(authBaseURL, 3*time.Second)
	}

	taskService := service.NewTaskService()

	mux := http.NewServeMux()
	handler := taskshttp.NewHandler(taskService, authVerifier)
	handler.RegisterRoutes(mux)

	httpHandler := middleware.RequestID(middleware.Logging(mux))

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      httpHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("Tasks HTTP server starting on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server stopped")
}
