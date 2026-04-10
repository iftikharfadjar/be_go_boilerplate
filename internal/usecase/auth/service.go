package auth

import (
	"context"

	"boilerplate/internal/domain"
)

type useCase struct {
	repo domain.AuthRepository
}

func NewAuthUseCase(repo domain.AuthRepository) domain.AuthUseCase {
	return &useCase{repo: repo}
}

func (u *useCase) Login(ctx context.Context, username, password string) (string, error) {
	// Business validation logic can go here
	_, token, err := u.repo.Login(ctx, username, password)
	return token, err
}

func (u *useCase) Signup(ctx context.Context, username, password string) (*domain.User, error) {
	// e.g. Check password strength, clean username, etc.
	return u.repo.Signup(ctx, username, password)
}

func (u *useCase) ValidateToken(ctx context.Context, token string) (*domain.User, error) {
	return u.repo.ValidateToken(ctx, token)
}
