package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/domain"
)

func (r *DashboardRepository) GetAnalytics(
	ctx context.Context,
	userID int,
	period domain.AnalyticsPeriod,
) (domain.Analytics, error) {
	if userID <= 0 {
		return domain.Analytics{}, fmt.Errorf("invalid user id")
	}

	now := time.Now().UTC()
	var from time.Time
	var bucket string

	switch period {
	case domain.AnalyticsPeriodDay:
		from = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		bucket = "toStartOfHour(created_at)"
	case domain.AnalyticsPeriodWeek:
		daysSinceMonday := (int(now.Weekday()) + 6) % 7
		monday := now.AddDate(0, 0, -daysSinceMonday)
		from = time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, time.UTC)
		bucket = "toStartOfDay(created_at)"
	case domain.AnalyticsPeriodMonth:
		from = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		bucket = "toStartOfDay(created_at)"
	case domain.AnalyticsPeriodYear:
		from = time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)
		bucket = "toStartOfMonth(created_at)"
	default:
		return domain.Analytics{}, fmt.Errorf("unsupported analytics period %q", period)
	}

	result := domain.Analytics{
		Period:        string(period),
		From:          from,
		To:            now,
		Series:        make([]domain.AnalyticsPoint, 0),
		TopCategories: make([]domain.CategoryStat, 0),
	}

	const summaryQuery = `
		SELECT
			toFloat64(sumIf(amount, type = 'income')) AS income,
			toFloat64(sumIf(amount, type = 'expense')) AS expenses
		FROM finance.transactions_analytics
		WHERE user_id = ? AND created_at >= ? AND created_at < ?`
	if err := r.conn.QueryRow(ctx, summaryQuery, userID, from, now).Scan(&result.Income, &result.Expenses); err != nil {
		return domain.Analytics{}, fmt.Errorf("query analytics summary: %w", err)
	}
	result.Net = result.Income - result.Expenses
	if result.Income > 0 {
		result.SavingsRate = result.Net / result.Income * 100
	}

	seriesQuery := fmt.Sprintf(`
		SELECT
			%s AS period,
			toFloat64(sumIf(amount, type = 'income')) AS income,
			toFloat64(sumIf(amount, type = 'expense')) AS expenses
		FROM finance.transactions_analytics
		WHERE user_id = ? AND created_at >= ? AND created_at < ?
		GROUP BY period
		ORDER BY period ASC`, bucket)

	rows, err := r.conn.Query(ctx, seriesQuery, userID, from, now)
	if err != nil {
		return domain.Analytics{}, fmt.Errorf("query analytics series: %w", err)
	}
	for rows.Next() {
		var point domain.AnalyticsPoint
		if err := rows.Scan(&point.Period, &point.Income, &point.Expenses); err != nil {
			rows.Close()
			return domain.Analytics{}, fmt.Errorf("scan analytics point: %w", err)
		}
		result.Series = append(result.Series, point)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return domain.Analytics{}, fmt.Errorf("iterate analytics series: %w", err)
	}
	rows.Close()

	const categoryQuery = `
		SELECT category, toFloat64(sum(amount)) AS total
		FROM finance.transactions_analytics
		WHERE user_id = ? AND type = 'expense' AND created_at >= ? AND created_at < ?
		GROUP BY category
		ORDER BY total DESC
		LIMIT 5`
	catRows, err := r.conn.Query(ctx, categoryQuery, userID, from, now)
	if err != nil {
		return domain.Analytics{}, fmt.Errorf("query analytics categories: %w", err)
	}
	for catRows.Next() {
		var category domain.CategoryStat
		if err := catRows.Scan(&category.Category, &category.Total); err != nil {
			catRows.Close()
			return domain.Analytics{}, fmt.Errorf("scan analytics category: %w", err)
		}
		if result.Expenses > 0 {
			category.Percentage = category.Total / result.Expenses * 100
		}
		result.TopCategories = append(result.TopCategories, category)
	}
	if err := catRows.Err(); err != nil {
		catRows.Close()
		return domain.Analytics{}, fmt.Errorf("iterate analytics categories: %w", err)
	}
	catRows.Close()

	return result, nil
}
