package grpc

import (
	"context"
	"log"

	pb "pz1.2/proto/auth"
	"pz1.2/services/auth/internal/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedAuthServiceServer
	authService *service.AuthService
}

func NewServer(authService *service.AuthService) *Server {
	return &Server{authService: authService}
}

func (s *Server) Verify(ctx context.Context, req *pb.VerifyRequest) (*pb.VerifyResponse, error) {
	log.Printf("[gRPC] Verify request for token: %s...", truncateToken(req.Token))

	resp, err := s.authService.Verify(req.Token)
	if err != nil {
		log.Printf("[gRPC] Token verification failed: %v", err)
		return &pb.VerifyResponse{
			Valid: false,
			Error: "unauthorized",
		}, status.Error(codes.Unauthenticated, "invalid token")
	}

	log.Printf("[gRPC] Token verified for subject: %s", resp.Subject)
	return &pb.VerifyResponse{
		Valid:   resp.Valid,
		Subject: resp.Subject,
	}, nil
}

func truncateToken(token string) string {
	if len(token) > 10 {
		return token[:10]
	}
	return token
}

func RegisterServer(s *grpc.Server, authService *service.AuthService) {
	pb.RegisterAuthServiceServer(s, NewServer(authService))
}
