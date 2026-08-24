package users

import (
	"backend/config"
	grpcclient "backend/internal/core/grpc"
	"backend/proto/users/gen"
	"fmt"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UsersClient struct {
	client gen.UserServiceClient
	conn   *grpc.ClientConn
}

func NewUsersClient(addr string, cfg *config.Config) (*UsersClient, error) {
	conn, err := grpcclient.NewGRPCClient(addr, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user service: %w", err)
	}
	return &UsersClient{
		client: gen.NewUserServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *UsersClient) Close() error {
	return c.conn.Close()
}

func (c *UsersClient) CreateProfile(ctx context.Context, req *CreateProfileRequest) (*UserProfile, error) {
	grpcReq := &gen.CreateProfileRequest{
		Email:        req.Email,
		FullName:     req.FullName,
		IsAdmin:      req.IsAdmin,
		PasswordHash: req.PasswordHash,
	}
	if req.PhoneNumber != nil {
		grpcReq.PhoneNumber = *req.PhoneNumber
	}

	resp, err := c.client.CreateProfile(ctx, grpcReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create profile: %w", err)
	}

	return &UserProfile{
		ID:          int(resp.Id),
		FullName:    req.FullName,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
		IsAdmin:     resp.IsAdmin,
		IsActive:    resp.IsActive,
	}, nil
}

func (c *UsersClient) ListUsers(ctx context.Context, req *gen.ListUsersRequest) (*ListUsersResponse, error) {
	grpcReq := &gen.ListUsersRequest{
		Offset: req.Offset,
		Limit:  req.Limit,
	}
	resp, err := c.client.ListUsers(ctx, grpcReq)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	users := make([]UserProfile, len(resp.Users))
	for i, user := range resp.Users {
		var phoneNumber *string
		if user.PhoneNumber != "" {
			phoneNumber = &user.PhoneNumber
		}
		users[i] = UserProfile{
			ID:          int(user.Id),
			FullName:    user.FullName,
			Email:       user.Email,
			PhoneNumber: phoneNumber,
			IsAdmin:     user.IsAdmin,
			IsActive:    user.IsActive,
		}
	}
	return &ListUsersResponse{
		Users:  users,
		Limit:  int(resp.Limit),
		Offset: int(resp.Offset),
	}, nil
}

func (c *UsersClient) DeleteUser(ctx context.Context, id int) error {
	grpcReq := &gen.DeleteUserRequest{
		Id: int32(id),
	}
	_, err := c.client.DeleteUser(ctx, grpcReq)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

func (c *UsersClient) UpdateRole(ctx context.Context, req *gen.UpdateRoleRequest) (*UserProfile, error) {
	grpcReq := &gen.UpdateRoleRequest{
		Id:      req.Id,
		IsAdmin: req.IsAdmin,
	}
	resp, err := c.client.UpdateRole(ctx, grpcReq)
	if err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}
	var phoneNumber *string
	if len(resp.PhoneNumber) != 0 && resp.PhoneNumber != "" {
		phoneNumber = &resp.PhoneNumber
	}
	return &UserProfile{
		ID:          int(resp.Id),
		FullName:    resp.FullName,
		Email:       resp.Email,
		PhoneNumber: phoneNumber,
		IsAdmin:     resp.IsAdmin,
		IsActive:    resp.IsActive,
	}, nil
}

func (c *UsersClient) GetUserByEmail(ctx context.Context, email string) (*UserProfile, error) {
	req := &gen.GetUserByEmailRequest{
		Email: email,
	}
	resp, err := c.client.GetUserByEmail(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	var phoneNumber *string
	if resp.PhoneNumber != "" {
		phoneNumber = &resp.PhoneNumber
	}
	return &UserProfile{
		ID:          int(resp.Id),
		FullName:    resp.FullName,
		Email:       resp.Email,
		PhoneNumber: phoneNumber,
		IsAdmin:     resp.IsAdmin,
		IsActive:    resp.IsActive,
	}, nil
}

func (c *UsersClient) GetMetrics(ctx context.Context) (*UserMetrics, error) {
	resp, err := c.client.GetMetrics(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("failed to get users metrics: %w", err)
	}
	return &UserMetrics{
		TotalUsers: int(resp.TotalUsers),
	}, nil
}

func (c *UsersClient) GetUser(ctx context.Context, req *gen.GetUserRequest) (*gen.UserResponse, error) {
	resp, err := c.client.GetUser(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return resp, nil
}

func (c *UsersClient) AdminExists(ctx context.Context) (bool, error) {
	resp, err := c.client.AdminExists(ctx, &emptypb.Empty{})
	if err != nil {
		return false, fmt.Errorf("failed to check admin exists: %w", err)
	}
	return resp.Exists, nil
}

func (c *UsersClient) MarkDeleting(ctx context.Context, id int) error {
	req := &gen.MarkDeletingRequest{Id: int32(id)}
	_, err := c.client.MarkDeleting(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to mark deleting: %w", err)
	}
	return nil
}

func (c *UsersClient) FinalizeDelete(ctx context.Context, id int) error {
	req := &gen.FinalizeDeleteRequest{Id: int32(id)}
	_, err := c.client.FinalizeDelete(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to finalize delete: %w", err)
	}
	return nil
}

func (c *UsersClient) RestoreUser(ctx context.Context, id int) error {
	req := &gen.RestoreUserRequest{Id: int32(id)}
	_, err := c.client.RestoreUser(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to restore user: %w", err)
	}
	return nil
}
