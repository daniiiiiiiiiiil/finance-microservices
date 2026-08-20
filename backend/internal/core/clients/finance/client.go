package finance

import (
	"backend/internal/features/finance/transport/grpc/proto"
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

var _ FinanceClientInterface = (*FinanceClient)(nil)

type FinanceClient struct {
	client proto.FinanceServiceClient
	conn   *grpc.ClientConn
}

func NewFinanceClient(addr string) (*FinanceClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to finance service: %w", err)
	}

	return &FinanceClient{
		client: proto.NewFinanceServiceClient(conn),
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
	req := &proto.DeleteUserTransactionsRequest{
		UserId: int32(userID),
	}

	_, err := c.client.DeleteUserTransaction(ctx, req)
	if err != nil {
		return fmt.Errorf("delete user transactions: %w", err)
	}
	return nil
}
