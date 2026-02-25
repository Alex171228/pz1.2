package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	authgrpc "pz1.2/services/auth/internal/grpc"
	authhttp "pz1.2/services/auth/internal/http"
	"pz1.2/services/auth/internal/service"
	"pz1.2/shared/middleware"

	"google.golang.org/grpc"
)

func main() {
	httpPort := os.Getenv("AUTH_PORT")
	if httpPort == "" {
		httpPort = "8081"
	}

	grpcPort := os.Getenv("AUTH_GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	authService := service.NewAuthService()

	mux := http.NewServeMux()
	handler := authhttp.NewHandler(authService)
	handler.RegisterRoutes(mux)

	httpHandler := middleware.RequestID(middleware.Logging(mux))

	httpServer := &http.Server{
		Addr:         ":" + httpPort,
		Handler:      httpHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	grpcServer := grpc.NewServer()
	authgrpc.RegisterServer(grpcServer, authService)

	go func() {
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			log.Fatalf("Failed to listen on gRPC port %s: %v", grpcPort, err)
		}
		log.Printf("Auth gRPC server starting on :%s", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	go func() {
		log.Printf("Auth HTTP server starting on :%s", httpPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down servers...")

	grpcServer.GracefulStop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("HTTP server shutdown failed: %v", err)
	}

	log.Println("Servers stopped")
}
