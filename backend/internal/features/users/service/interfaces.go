package service_user

import (
	"backend/internal/core/domain"
	"backend/internal/features/users/transport/grpc/proto"
	"context"
)

type UsersServiceInterface interface {
	CreateProfile(ctx context.Context, req *proto.CreateProfileRequest) (domain.User, error)
	GetUser(ctx context.Context, id int) (domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
	ListUsers(ctx context.Context, limit, offset int) ([]domain.User, int, error)
	PatchUser(ctx context.Context, id int, patch domain.UserPatch) (domain.User, error)
	DeleteUser(ctx context.Context, id int) error
	UpdateRole(ctx context.Context, id int, isAdmin bool) (domain.User, error)
	GetMetrics(ctx context.Context) (int, error)
	AdminExists(ctx context.Context) (bool, error)
	MarkDeleting(ctx context.Context, id int) error
	FinalizeDelete(ctx context.Context, id int) error
	RestoreUser(ctx context.Context, id int) error
}
