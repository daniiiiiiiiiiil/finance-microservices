package postgres

import (
	"backend/internal/core/domain"
	"context"
	"fmt"
	"time"
)

func (r *FinanceRepository) GetDashboard(ctx context.Context, userID int) (domain.Dashboard, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	var dashboard domain.Dashboard

	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	err := r.pool.QueryRow(ctx, `
		SELECT 
			COALESCE(SUM(CASE WHEN type_transaction = 'income' THEN amount ELSE -amount END), 0) as balance,
			COALESCE(SUM(CASE WHEN type_transaction = 'income' THEN amount ELSE 0 END), 0) as income,
			COALESCE(SUM(CASE WHEN type_transaction = 'expense' THEN amount ELSE 0 END), 0) as expense
		FROM finance.transactions
		WHERE user_id = $1 AND created_at >= $2
	`, userID, startOfMonth).Scan(
		&dashboard.TotalBalance,
		&dashboard.MonthlyIncome,
		&dashboard.MonthlyExpenses,
	)
	if err != nil {
		return domain.Dashboard{}, fmt.Errorf("get metrics: %w", err)
	}

	if dashboard.MonthlyIncome > 0 {
		dashboard.SavingsRate = (dashboard.MonthlyIncome - dashboard.MonthlyExpenses) / dashboard.MonthlyIncome * 100
		if dashboard.SavingsRate < 0 {
			dashboard.SavingsRate = 0
		}
	}

	rows, err := r.pool.Query(ctx, `
		SELECT 
			DATE(created_at) as date,
			COALESCE(SUM(CASE WHEN type_transaction = 'income' THEN amount ELSE 0 END), 0) as income,
			COALESCE(SUM(CASE WHEN type_transaction = 'expense' THEN amount ELSE 0 END), 0) as expense
		FROM finance.transactions
		WHERE user_id = $1 AND created_at >= $2
		GROUP BY DATE(created_at)
		ORDER BY date ASC
	`, userID, startOfMonth)
	if err != nil {
		return domain.Dashboard{}, fmt.Errorf("get daily stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var stat domain.DailyStat
		if err := rows.Scan(&stat.Date, &stat.Income, &stat.Expense); err != nil {
			return domain.Dashboard{}, fmt.Errorf("scan daily stat: %w", err)
		}
		dashboard.DailyStats = append(dashboard.DailyStats, stat)
	}

	rows, err = r.pool.Query(ctx, `
		SELECT id, type_transaction, amount, category, created_at
		FROM finance.transactions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 10
	`, userID)
	if err != nil {
		return domain.Dashboard{}, fmt.Errorf("get recent transactions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tx domain.RecentTransaction
		if err := rows.Scan(&tx.ID, &tx.Type, &tx.Amount, &tx.Category, &tx.CreatedAt); err != nil {
			return domain.Dashboard{}, fmt.Errorf("scan transaction: %w", err)
		}
		dashboard.RecentTxs = append(dashboard.RecentTxs, tx)
	}

	rows, err = r.pool.Query(ctx, `
		SELECT category, SUM(amount) as total
		FROM finance.transactions
		WHERE user_id = $1 AND type_transaction = 'expense' AND created_at >= $2
		GROUP BY category
		ORDER BY total DESC
		LIMIT 5
	`, userID, startOfMonth)
	if err != nil {
		return domain.Dashboard{}, fmt.Errorf("get top categories: %w", err)
	}
	defer rows.Close()

	var totalExpenses float64
	for _, stat := range dashboard.DailyStats {
		totalExpenses += stat.Expense
	}

	for rows.Next() {
		var cat domain.CategoryStat
		if err := rows.Scan(&cat.Category, &cat.Total); err != nil {
			return domain.Dashboard{}, fmt.Errorf("scan category: %w", err)
		}
		if totalExpenses > 0 {
			cat.Percentage = (cat.Total / totalExpenses) * 100
		}
		dashboard.TopCategories = append(dashboard.TopCategories, cat)
	}

	return dashboard, nil
}
