package gRPC

import (
	"backend/internal/core/domain"
	"backend/internal/features/finance/transport/gRPC/proto"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func convertFinanceToProto(finance domain.Finance) *proto.TransactionResponse {
	return &proto.TransactionResponse{
		Id:              int32(finance.ID),
		Version:         int32(finance.Version),
		TypeTransaction: finance.TypeTransaction,
		Amount:          finance.Amount,
		Category:        finance.Category,
		CreatedAt:       timestamppb.New(finance.CreatedAt),
		UserId:          int32(finance.UserID),
	}
}

func ConvertFinanceToProto(finance []domain.Finance) []*proto.TransactionResponse {
	result := make([]*proto.TransactionResponse, len(finance))
	for i, fin := range finance {
		result[i] = convertFinanceToProto(fin)
	}
	return result
}

func convertDashboardToProto(dashboard domain.Dashboard) *proto.DashboardResponse {
	dailyStats := make([]*proto.DailyStat, len(dashboard.DailyStats))
	for i, stat := range dashboard.DailyStats {
		dailyStats[i] = &proto.DailyStat{
			Date:    stat.Date.Format("2006-01-02"),
			Income:  stat.Income,
			Expense: stat.Expense,
		}
	}

	recentTxs := make([]*proto.RecentTransaction, len(dashboard.RecentTxs))
	for i, tx := range dashboard.RecentTxs {
		recentTxs[i] = &proto.RecentTransaction{
			Id:        int32(tx.ID),
			Type:      tx.Type,
			Amount:    tx.Amount,
			Category:  tx.Category,
			CreatedAt: timestamppb.New(tx.CreatedAt),
		}
	}

	topCategories := make([]*proto.CategoryStat, len(dashboard.TopCategories))
	for i, cat := range dashboard.TopCategories {
		topCategories[i] = &proto.CategoryStat{
			Category:   cat.Category,
			Total:      cat.Total,
			Percentage: cat.Percentage,
		}
	}

	return &proto.DashboardResponse{
		TotalBalance:       dashboard.TotalBalance,
		MonthlyIncome:      dashboard.MonthlyIncome,
		MonthlyExpenses:    dashboard.MonthlyExpenses,
		SavingsRate:        dashboard.SavingsRate,
		DailyStats:         dailyStats,
		RecentTransactions: recentTxs,
		TopCategories:      topCategories,
	}
}
