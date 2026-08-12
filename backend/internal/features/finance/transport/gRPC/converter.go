package gRPC

import (
	"backend/internal/core/domain"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func convertFinanceToProto(finance domain.Finance) *TransactionResponse {
	return &TransactionResponse{
		Id:              int32(finance.ID),
		Version:         int32(finance.Version),
		TypeTransaction: finance.TypeTransaction,
		Amount:          finance.Amount,
		Category:        finance.Category,
		CreatedAt:       timestamppb.New(finance.CreatedAt),
		UserId:          int32(finance.UserID),
	}
}

func ConvertFinanceToProto(finance []domain.Finance) []*TransactionResponse {
	result := make([]*TransactionResponse, len(finance))
	for i, fin := range finance {
		result[i] = convertFinanceToProto(fin)
	}
	return result
}

func convertDashboardToProto(dashboard domain.Dashboard) *DashboardResponse {
	dailyStats := make([]*DailyStat, len(dashboard.DailyStats))
	for i, stat := range dashboard.DailyStats {
		dailyStats[i] = &DailyStat{
			Date:    stat.Date.Format("2006-01-02"),
			Income:  stat.Income,
			Expense: stat.Expense,
		}
	}

	recentTxs := make([]*RecentTransaction, len(dashboard.RecentTxs))
	for i, tx := range dashboard.RecentTxs {
		recentTxs[i] = &RecentTransaction{
			Id:        int32(tx.ID),
			Type:      tx.Type,
			Amount:    tx.Amount,
			Category:  tx.Category,
			CreatedAt: timestamppb.New(tx.CreatedAt),
		}
	}

	topCategories := make([]*CategoryStat, len(dashboard.TopCategories))
	for i, cat := range dashboard.TopCategories {
		topCategories[i] = &CategoryStat{
			Category:   cat.Category,
			Total:      cat.Total,
			Percentage: cat.Percentage,
		}
	}

	return &DashboardResponse{
		TotalBalance:       dashboard.TotalBalance,
		MonthlyIncome:      dashboard.MonthlyIncome,
		MonthlyExpenses:    dashboard.MonthlyExpenses,
		SavingsRate:        dashboard.SavingsRate,
		DailyStats:         dailyStats,
		RecentTransactions: recentTxs,
		TopCategories:      topCategories,
	}
}
