package grpc

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/proto/finance/gen"
)

func convertFinanceToProto(finance domain.Finance) *gen.TransactionResponse {
	return &gen.TransactionResponse{
		Id:              int32(finance.ID),
		Version:         int32(finance.Version),
		TypeTransaction: finance.TypeTransaction,
		Amount:          finance.Amount,
		Category:        finance.Category,
		CreatedAt:       timestamppb.New(finance.CreatedAt),
		UserId:          int32(finance.UserID),
	}
}

func convertDashboardToProto(dashboard domain.Dashboard) *gen.DashboardResponse {
	dailyStats := make([]*gen.DailyStat, len(dashboard.DailyStats))
	for i, stat := range dashboard.DailyStats {
		dailyStats[i] = &gen.DailyStat{
			Date:    stat.Date.Format("2006-01-02"),
			Income:  stat.Income,
			Expense: stat.Expense,
		}
	}

	recentTxs := make([]*gen.RecentTransaction, len(dashboard.RecentTxs))
	for i, tx := range dashboard.RecentTxs {
		recentTxs[i] = &gen.RecentTransaction{
			Id:        int32(tx.ID),
			Type:      tx.Type,
			Amount:    tx.Amount,
			Category:  tx.Category,
			CreatedAt: timestamppb.New(tx.CreatedAt),
		}
	}

	topCategories := make([]*gen.CategoryStat, len(dashboard.TopCategories))
	for i, cat := range dashboard.TopCategories {
		topCategories[i] = &gen.CategoryStat{
			Category:   cat.Category,
			Total:      cat.Total,
			Percentage: cat.Percentage,
		}
	}

	return &gen.DashboardResponse{
		TotalBalance:       dashboard.TotalBalance,
		MonthlyIncome:      dashboard.MonthlyIncome,
		MonthlyExpenses:    dashboard.MonthlyExpenses,
		SavingsRate:        dashboard.SavingsRate,
		DailyStats:         dailyStats,
		RecentTransactions: recentTxs,
		TopCategories:      topCategories,
	}
}

func convertAnalyticsToProto(analytics domain.Analytics) *gen.AnalyticsResponse {
	series := make([]*gen.AnalyticsPoint, 0, len(analytics.Series))
	for _, point := range analytics.Series {
		series = append(series, &gen.AnalyticsPoint{
			Period:   timestamppb.New(point.Period),
			Income:   point.Income,
			Expenses: point.Expenses,
		})
	}

	categories := make([]*gen.CategoryStat, 0, len(analytics.TopCategories))
	for _, category := range analytics.TopCategories {
		categories = append(categories, &gen.CategoryStat{
			Category:   category.Category,
			Total:      category.Total,
			Percentage: category.Percentage,
		})
	}

	return &gen.AnalyticsResponse{
		Period:        analytics.Period,
		From:          timestamppb.New(analytics.From),
		To:            timestamppb.New(analytics.To),
		Income:        analytics.Income,
		Expenses:      analytics.Expenses,
		Net:           analytics.Net,
		SavingsRate:   analytics.SavingsRate,
		Series:        series,
		TopCategories: categories,
	}
}

func convertBudgetToProto(b domain.Budget) *gen.Budget {
	return &gen.Budget{
		Id:          int32(b.ID),
		Version:     int32(b.Version),
		UserId:      int32(b.UserID),
		Category:    b.Category,
		AmountLimit: b.AmountLimit,
		Period:      string(b.Period),
		PeriodStart: timestamppb.New(b.PeriodStart),
		PeriodEnd:   timestamppb.New(b.PeriodEnd),
	}
}

func convertBudgetStatusToProto(st domain.BudgetStatus) *gen.BudgetStatus {
	return &gen.BudgetStatus{
		Budget:     convertBudgetToProto(st.Budget),
		Spent:      st.Spent,
		Remaining:  st.Remaining,
		Percentage: st.Percentage,
		Exceeded:   st.Exceeded,
	}
}
