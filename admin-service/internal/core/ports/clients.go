package ports

import (
	"context"

	financeClient "github.com/daniiiiiiiiiiil/finance-microservices/admin-service/internal/core/clients/finance"
	"github.com/daniiiiiiiiiiil/finance-microservices/admin-service/internal/core/clients/users"
	userpb "github.com/daniiiiiiiiiiil/finance-microservices/admin-service/proto/users/gen"
)

type UsersClientInterface interface {
	GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.UserResponse, error)
	ListUsers(ctx context.Context, req *userpb.ListUsersRequest) (*users.ListUsersResponse, error)
	UpdateRole(ctx context.Context, req *userpb.UpdateRoleRequest) (*users.UserProfile, error)
	MarkDeleting(ctx context.Context, id int) error
	RestoreUser(ctx context.Context, id int) error
	FinalizeDelete(ctx context.Context, id int) error
	GetMetrics(ctx context.Context) (*users.UserMetrics, error)
	Close() error
}

type FinanceClientInterface interface {
	DeleteUserTransactions(ctx context.Context, userID int) error
	GetMetrics(ctx context.Context) (*financeClient.FinanceMetrics, error)
	Close() error
}
