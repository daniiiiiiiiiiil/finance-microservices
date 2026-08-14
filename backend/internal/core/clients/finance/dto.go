package finance

import "context"

type FinanceClientInterface interface {
	GetMetrics(ctx context.Context) (*FinanceMetrics, error)
	Close() error
}

type FinanceMetrics struct {
	TotalTransactions int     `json:"total_transactions"`
	TotalBalance      float64 `json:"total_balance"`
}
