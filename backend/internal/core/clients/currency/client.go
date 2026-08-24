package currency

import (
	"backend/config"
	grpcclient "backend/internal/core/grpc"
	"backend/proto/currency/gen"
	"fmt"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type CurrencyClient struct {
	client gen.CurrencyServiceClient
	conn   *grpc.ClientConn
}

func NewCurrencyClient(addr string, cfg *config.Config) (*CurrencyClient, error) {
	conn, err := grpcclient.NewGRPCClient(addr, cfg)
	if err != nil {
		return nil, fmt.Errorf("could not connect to CurrencyServer %s", addr)
	}
	return &CurrencyClient{
		client: gen.NewCurrencyServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *CurrencyClient) GetRates(ctx context.Context, base string) (*gen.GetRatesResponse, error) {
	resp, err := c.client.GetRates(ctx, &gen.GetRatesRequest{Base: base})
	if err != nil {
		return nil, fmt.Errorf("could not get rates for base %s: %w", base, err)
	}
	return resp, nil
}

func (c *CurrencyClient) Convert(ctx context.Context, from string, to string, amount float64) (float64, error) {
	resp, err := c.client.Convert(ctx, &gen.ConvertRequest{
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
