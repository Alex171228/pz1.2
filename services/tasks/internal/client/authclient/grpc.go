package authclient

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "pz1.2/proto/auth"
	"pz1.2/shared/middleware"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type GRPCClient struct {
	conn    *grpc.ClientConn
	client  pb.AuthServiceClient
	timeout time.Duration
}

func NewGRPCClient(addr string, timeout time.Duration) (*GRPCClient, error) {
	conn, err := grpc.Dial(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("connect to auth service: %w", err)
	}

	return &GRPCClient{
		conn:    conn,
		client:  pb.NewAuthServiceClient(conn),
		timeout: timeout,
	}, nil
}

func (c *GRPCClient) Verify(ctx context.Context, token string) (*VerifyResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	requestID := middleware.GetRequestID(ctx)
	log.Printf("[%s] Calling Auth gRPC verify", requestID)

	resp, err := c.client.Verify(ctx, &pb.VerifyRequest{Token: token})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.Unauthenticated {
			log.Printf("[%s] Auth gRPC verify: unauthorized", requestID)
			return &VerifyResponse{
				Valid: false,
				Error: "unauthorized",
			}, nil
		}
		log.Printf("[%s] Auth gRPC verify failed: %v", requestID, err)
		return nil, fmt.Errorf("auth service error: %w", err)
	}

	log.Printf("[%s] Auth gRPC verify: success, subject=%s", requestID, resp.Subject)
	return &VerifyResponse{
		Valid:   resp.Valid,
		Subject: resp.Subject,
	}, nil
}

func (c *GRPCClient) Close() error {
	return c.conn.Close()
}
