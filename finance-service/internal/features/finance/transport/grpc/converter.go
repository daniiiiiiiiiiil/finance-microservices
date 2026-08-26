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

func ConvertFinanceToProto(finance []domain.Finance) []*gen.TransactionResponse {
	result := make([]*gen.TransactionResponse, len(finance))
	for i, fin := range finance {
		result[i] = convertFinanceToProto(fin)
	}
	return result
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
