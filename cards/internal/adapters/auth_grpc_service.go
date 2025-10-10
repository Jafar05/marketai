package adapters

import (
	"context"
	"fmt"
	auth_grpc_api "marketai/auth/proto/generated-source"
	"marketai/cards/internal/config"
	"marketai/cards/internal/domain"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthGRPCService struct {
	client auth_grpc_api.AuthServiceClient
	conn   *grpc.ClientConn
}

func NewAuthGRPCService(cfg *config.Config) (*AuthGRPCService, error) {
	endpoint := cfg.Auth.GRPCEndpoint
	conn, err := grpc.Dial(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth service: %w", err)
	}

	client := auth_grpc_api.NewAuthServiceClient(conn)

	return &AuthGRPCService{
		client: client,
		conn:   conn,
	}, nil
}

func (s *AuthGRPCService) ValidateToken(ctx context.Context, token string) (*domain.UserInfo, error) {
	req := &auth_grpc_api.ValidateTokenRequest{
		Token: token,
	}

	resp, err := s.client.ValidateToken(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}

	if !resp.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return &domain.UserInfo{
		UserID: resp.UserId,
		Role:   resp.Role,
	}, nil
}

func (s *AuthGRPCService) Close() error {
	return s.conn.Close()
}
