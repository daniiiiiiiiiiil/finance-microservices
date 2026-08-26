package finance

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/daniiiiiiiiiiil/finance-microservices/admin-service/config"
	grpcclient "github.com/daniiiiiiiiiiil/finance-microservices/admin-service/internal/core/grpc"
	"github.com/daniiiiiiiiiiil/finance-microservices/admin-service/proto/finance/gen"
)

var _ FinanceClientInterface = (*FinanceClient)(nil)

type FinanceClient struct {
	client gen.FinanceServiceClient
	conn   *grpc.ClientConn
}

func NewFinanceClient(addr string, cfg *config.Config) (*FinanceClient, error) {
	conn, err := grpcclient.NewGRPCClient(addr, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to finance service: %w", err)
	}

	return &FinanceClient{
		client: gen.NewFinanceServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *FinanceClient) Close() error {
	return c.conn.Close()
}

func (c *FinanceClient) GetMetrics(ctx context.Context) (*FinanceMetrics, error) {
	resp, err := c.client.GetMetrics(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("failed to get finance metrics: %w", err)
	}

	return &FinanceMetrics{
		TotalTransactions: int(resp.TotalTransactions),
		TotalBalance:      resp.TotalBalance,
	}, nil
}

func (c *FinanceClient) DeleteUserTransactions(ctx context.Context, userID int) error {
	req := &gen.DeleteUserTransactionsRequest{
		UserId: int32(userID),
	}

	_, err := c.client.DeleteUserTransaction(ctx, req)
	if err != nil {
		return fmt.Errorf("delete user transactions: %w", err)
	}
	return nil
}
