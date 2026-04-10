package pocketbase

import (
	"context"
	"errors"

	"github.com/pocketbase/pocketbase"
	"boilerplate/internal/domain"
)

type authRepository struct {
	pb *pocketbase.PocketBase
}

func NewAuthRepository(pb *pocketbase.PocketBase) domain.AuthRepository {
	return &authRepository{pb: pb}
}

func (r *authRepository) Login(ctx context.Context, username, password string) (*domain.User, string, error) {
	// In a real implementation this would use r.pb.FindAuthRecordByUsername() or similar
	// For boilerplate we return stub data
	return &domain.User{ID: "stub-id", Username: username}, "stub-jwt-token", nil
}

func (r *authRepository) Signup(ctx context.Context, username, password string) (*domain.User, error) {
	// Real implementation creates record in PB
	return &domain.User{ID: "new-id", Username: username}, nil
}

func (r *authRepository) ValidateToken(ctx context.Context, token string) (*domain.User, error) {
	// Real implementation uses CheckAuthToken
	if token == "" {
		return nil, errors.New("empty token")
	}
	return &domain.User{ID: "valid-id", Username: "validated-user"}, nil
}
