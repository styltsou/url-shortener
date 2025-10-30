package service

import (
	"context"

	"github.com/styltsou/url-shortener/server/pkg/db"
)

type UserService struct {
	queries *db.Queries
}

func NewUserService(queries *db.Queries) *UserService {
	return &UserService{queries: queries}
}

// * These 2 methods are sufficient for now

// TODO: See if i also need to add first name and last name
func (s *UserService) CreateUser(ctx context.Context, clerkID string, email string, avatarUrl *string) (db.User, error) {
	return s.queries.CreateUser(ctx, db.CreateUserParams{
		ClerkID:   clerkID,
		Email:     email,
		AvatarUrl: avatarUrl,
	})
}

func (s *UserService) UpdateUser(ctx context.Context) {}
