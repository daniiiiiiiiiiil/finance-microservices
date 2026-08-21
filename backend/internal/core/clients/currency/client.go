package currency

import (
	"backend/internal/features/currency/transport/grpc/proto"
	"fmt"

	"golang.org/x/net/context"
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
		return nil, fmt.Errorf("could not connect to CurrencyServer %s", addr)
	}
	return &CurrencyClient{
		client: proto.NewCurrencyServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *CurrencyClient) GetRates(ctx context.Context, base string) (*proto.GetRatesResponse, error) {
	resp, err := c.client.GetRates(ctx, &proto.GetRatesRequest{Base: base})
	if err != nil {
		return nil, fmt.Errorf("could not get rates for base %s: %w", base, err)
	}
	return resp, nil
}

func (c *CurrencyClient) Convert(ctx context.Context, from string, to string, amount float64) (float64, error) {
	resp, err := c.client.Convert(ctx, &proto.ConvertRequest{
		From:   from,
		To:     to,
		Amount: amount,
	})
	if err != nil {
		return 0, fmt.Errorf("could not convert %s to %s: %w", from, to, err)
	}
	return resp.Result, nil
}

func (c *CurrencyClient) Close() error {
	return c.conn.Close()
}
