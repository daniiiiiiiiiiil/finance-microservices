package currency

import (
	"backend/internal/features/currency/transport/grpc/proto"
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type CurrencyClient struct {
	client proto.CurrencyServiceClient
	conn   *grpc.ClientConn
}

func NewCurrencyClient(addr string) (*CurrencyClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to currency service: %w", err)
	}

	return &CurrencyClient{
		client: proto.NewCurrencyServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *CurrencyClient) GetRates(ctx context.Context, base string) (*proto.GetRatesResponse, error) {
	req := &proto.GetRatesRequest{Base: base}
	return c.client.GetRates(ctx, req)
}

func (c *CurrencyClient) Convert(ctx context.Context, from, to string, amount float64) (*proto.ConvertResponse, error) {
	req := &proto.ConvertRequest{
		From:   from,
		To:     to,
		Amount: amount,
	}
	return c.client.Convert(ctx, req)
}

func (c *CurrencyClient) GetTransactionUSD(ctx context.Context, txID int64) (*proto.GetTransactionUSDResponse, error) {
	req := &proto.GetTransactionUSDRequest{Id: txID}
	return c.client.GetTransactionUSD(ctx, req)
}

func (c *CurrencyClient) Close() error {
	return c.conn.Close()
}
