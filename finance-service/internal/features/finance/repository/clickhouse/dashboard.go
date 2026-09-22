package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/domain"
)

type DashboardRepository struct {
	conn driver.Conn
}

func NewDashboardRepository(conn driver.Conn) *DashboardRepository {
	return &DashboardRepository{conn: conn}
}

func (r *DashboardRepository) GetDashboard(ctx context.Context, userID int) (domain.Dashboard, error) {
	var dashboard domain.Dashboard

	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	query := `
		SELECT
			toFloat64(COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE -amount END), 0)) AS balance,
			toFloat64(COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0)) AS income,
			toFloat64(COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0)) AS expense
		FROM finance.transactions_analytics
		WHERE user_id = ? AND created_at >= ?
	`
	err := r.conn.QueryRow(ctx, query, userID, startOfMonth).Scan(
		&dashboard.TotalBalance,
		&dashboard.MonthlyIncome,
		&dashboard.MonthlyExpenses,
	)
	if err != nil {
		return domain.Dashboard{}, fmt.Errorf("get dashboard failed: %w", err)
	}

	if dashboard.MonthlyIncome > 0 {
		dashboard.SavingsRate = (dashboard.MonthlyIncome - dashboard.MonthlyExpenses) / dashboard.MonthlyIncome * 100
		if dashboard.SavingsRate < 0 {
			dashboard.SavingsRate = 0
		}
	}

	queryRows := `
		SELECT
			toDate(created_at) AS date,
			toFloat64(COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0)) AS income,
			toFloat64(COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0)) AS expense
		FROM finance.transactions_analytics
		WHERE user_id = ? AND created_at >= ?
		GROUP BY date
		ORDER BY date ASC
	`
	rows, err := r.conn.Query(ctx, queryRows, userID, startOfMonth)
	if err != nil {
		return domain.Dashboard{}, fmt.Errorf("get daily stats: %w", err)
	}
	for rows.Next() {
		var stat domain.DailyStat
		if err := rows.Scan(&stat.Date, &stat.Income, &stat.Expense); err != nil {
			rows.Close()
			return domain.Dashboard{}, fmt.Errorf("scan daily stat: %w", err)
		}
		dashboard.DailyStats = append(dashboard.DailyStats, stat)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return domain.Dashboard{}, fmt.Errorf("daily stats rows error: %w", err)
	}
	rows.Close()

	queryTransaction := `
		SELECT transaction_id, type, toFloat64(amount), category, created_at
		FROM finance.transactions_analytics
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT 10
	`
	rows, err = r.conn.Query(ctx, queryTransaction, userID)
	if err != nil {
		return domain.Dashboard{}, fmt.Errorf("get recent transactions: %w", err)
	}
	for rows.Next() {
		var tx domain.RecentTransaction
		if err := rows.Scan(&tx.ID, &tx.Type, &tx.Amount, &tx.Category, &tx.CreatedAt); err != nil {
			rows.Close()
			return domain.Dashboard{}, fmt.Errorf("scan transaction: %w", err)
		}
		dashboard.RecentTxs = append(dashboard.RecentTxs, tx)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return domain.Dashboard{}, fmt.Errorf("recent transactions rows error: %w", err)
	}
	rows.Close()

	var totalExpenses float64
	for _, stat := range dashboard.DailyStats {
		totalExpenses += stat.Expense
	}

	queryTop := `
		SELECT category, toFloat64(SUM(amount)) AS total
		FROM finance.transactions_analytics
		WHERE user_id = ? AND type = 'expense' AND created_at >= ?
		GROUP BY category
		ORDER BY total DESC
		LIMIT 5
	`
	rows, err = r.conn.Query(ctx, queryTop, userID, startOfMonth)
	if err != nil {
		return domain.Dashboard{}, fmt.Errorf("get top categories: %w", err)
	}
	for rows.Next() {
		var cat domain.CategoryStat
		if err := rows.Scan(&cat.Category, &cat.Total); err != nil {
			rows.Close()
			return domain.Dashboard{}, fmt.Errorf("scan category: %w", err)
		}
		if totalExpenses > 0 {
			cat.Percentage = (cat.Total / totalExpenses) * 100
		}
		dashboard.TopCategories = append(dashboard.TopCategories, cat)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return domain.Dashboard{}, fmt.Errorf("top categories rows error: %w", err)
	}
	rows.Close()

	return dashboard, nil
}

func (r *DashboardRepository) SaveTransaction(ctx context.Context, tx domain.Finance) error {
	query := `
		INSERT INTO finance.transactions_analytics
		(transaction_id, user_id, type, amount, category, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	err := r.conn.Exec(ctx, query,
		tx.ID,
		tx.UserID,
		tx.TypeTransaction,
		tx.Amount,
		tx.Category,
		tx.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("save transaction failed: %w", err)
	}
	return nil
}

func (r *DashboardRepository) DeleteTransaction(ctx context.Context, transactionID int) error {
	query := `ALTER TABLE finance.transactions_analytics DELETE WHERE transaction_id = ?`
	err := r.conn.Exec(ctx, query, transactionID)
	if err != nil {
		return fmt.Errorf("delete transaction failed: %w", err)
	}
	return nil
}
