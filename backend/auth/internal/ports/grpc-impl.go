package ports

import (
	"context"
	"marketai/auth/internal/app"
	auth_grpc_api "marketai/auth/proto/generated-source"
)

type grpcServiceImpl struct {
	appCQRS *app.AppCQRS
	auth_grpc_api.UnimplementedAuthServiceServer
}

func newGrpcServer(appCQRS *app.AppCQRS) auth_grpc_api.AuthServiceServer {
	return &grpcServiceImpl{
		appCQRS: appCQRS,
	}
}

func (s *grpcServiceImpl) GetUserData(
	ctx context.Context,
	req *auth_grpc_api.GetUserDataRequest,
) (*auth_grpc_api.GetUserDataResponse, error) {
	user, err := s.appCQRS.Queries.GetUserByToken.Handle(ctx, req.Token)
	if err != nil {
		return nil, err
	}

	return &auth_grpc_api.GetUserDataResponse{
		Email: user.Email,
	}, nil
}
