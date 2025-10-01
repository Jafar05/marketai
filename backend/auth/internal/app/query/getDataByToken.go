package query

import (
	"context"
	"marketai/auth/internal/domain"
)

type GetDataByTokenHandler interface {
	Handle(ctx context.Context, token string) (*domain.GetData, error)
}

type getDataByTokenHandler struct {
	repo domain.UserRepository
}

func NewGetDataByTokenHandler(repo domain.UserRepository) GetDataByTokenHandler {
	return &getDataByTokenHandler{
		repo: repo,
	}
}

func (r *getDataByTokenHandler) Handle(ctx context.Context, token string) (*domain.GetData, error) {
	return r.repo.GetDataByToken(ctx, token)
}
