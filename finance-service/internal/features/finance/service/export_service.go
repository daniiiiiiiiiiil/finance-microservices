package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/ports"
	"github.com/jung-kurt/gofpdf"
)

type ExportService struct {
	repo ports.FinanceRepositoryInterface
	s3   ports.S3Interface
}

func NewExportService(repo ports.FinanceRepositoryInterface, s3 ports.S3Interface) *ExportService {
	return &ExportService{repo: repo, s3: s3}
}

func (s *ExportService) ExportJSON(ctx context.Context, userID int) (string, error) {
	transactions, err := s.repo.GetTransactions(ctx, userID, nil, nil, nil, nil, 100000, 0)
	if err != nil {
		return "", fmt.Errorf("failed to get transactions: %w", err)
	}

	exportData := map[string]interface{}{
		"user_id":      userID,
		"exported_at":  time.Now().Format(time.RFC3339),
		"transactions": transactions,
	}

	data, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal exported data: %w", err)
	}

	key := fmt.Sprintf("exports/user_%d/export_%s.json", userID, time.Now().Format("2006-01-02"))

	if err := s.s3.Put(ctx, key, data); err != nil {
		return "", fmt.Errorf("failed to put exported data: %w", err)
	}
	return key, nil
}

func (s *ExportService) ExportCSV(ctx context.Context, userID int) (string, error) {
	transactions, err := s.repo.GetTransactions(ctx, userID, nil, nil, nil, nil, 10000, 0)
	if err != nil {
		return "", fmt.Errorf("failed to get transactions: %w", err)
	}
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	if err := writer.Write([]string{"ID", "Type", "Amount", "Category", "Date"}); err != nil {
		return "", fmt.Errorf("failed to write CSV header: %w", err)
	}

	for _, tx := range transactions {
		err = writer.Write([]string{
			fmt.Sprintf("%d", tx.ID),
			tx.TypeTransaction,
			fmt.Sprintf("%f", tx.Amount),
			tx.Category,
			tx.CreatedAt.Format("2006-01-02 15:04:05"),
		})
		if err != nil {
			return "", fmt.Errorf("failed to write CSV row: %w", err)
		}
	}
	writer.Flush()

	key := fmt.Sprintf("exports/user_%d/export_%s.csv", userID, time.Now().Format("2006-01-02"))
	if err := s.s3.Put(ctx, key, buf.Bytes()); err != nil {
		return "", fmt.Errorf("failed to put exported data: %w", err)
	}
	return key, nil
}

func (s *ExportService) ExportTXT(ctx context.Context, userID int) (string, error) {
	transactions, err := s.repo.GetTransactions(ctx, userID, nil, nil, nil, nil, 10000, 0)
	if err != nil {
		return "", fmt.Errorf("failed to get transactions: %w", err)
	}

	var report bytes.Buffer
	report.WriteString("=== ФИНАНСОВЫЙ ОТЧЕТ ===\n")
	report.WriteString(fmt.Sprintf("Пользователь: %d\n", userID))
	report.WriteString(fmt.Sprintf("Дата: %s\n", time.Now().Format("2006-01-02")))
	report.WriteString("\nТранзакции:\n")
	report.WriteString("--------------------------------\n")

	for _, tx := range transactions {
		report.WriteString(fmt.Sprintf(
			"%s | %s | %.2f ₽ | %s\n",
			tx.CreatedAt.Format("2006-01-02"),
			tx.TypeTransaction,
			tx.Amount,
			tx.Category,
		))
	}

	key := fmt.Sprintf("exports/user_%d/export_%s.txt", userID, time.Now().Format("2006-01-02"))
	if err := s.s3.Put(ctx, key, report.Bytes()); err != nil {
		return "", fmt.Errorf("failed to save to S3: %w", err)
	}

	return key, nil
}

func (s *ExportService) GetExportURL(ctx context.Context, key string) (string, error) {
	return fmt.Sprintf("http://localhost:9000/finance-files/%s", key), nil
}

func (s *ExportService) ExportPDF(ctx context.Context, userID int) (string, error) {
	transactions, err := s.repo.GetTransactions(ctx, userID, nil, nil, nil, nil, 10000, 0)
	if err != nil {
		return "", fmt.Errorf("failed to get transactions: %w", err)
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Helvetica", "", 16)
	pdf.Cell(40, 10, "Financial Report")
	pdf.Ln(10)

	pdf.SetFont("Helvetica", "", 12)
	pdf.Cell(40, 10, fmt.Sprintf("User: %d", userID))
	pdf.Ln(5)
	pdf.Cell(40, 10, fmt.Sprintf("Date: %s", time.Now().Format("2006-01-02")))
	pdf.Ln(10)

	pdf.SetFont("Helvetica", "B", 12)
	pdf.Cell(40, 10, "Transactions:")
	pdf.Ln(5)

	for _, tx := range transactions {
		pdf.SetFont("Helvetica", "", 10)
		pdf.Cell(40, 10, fmt.Sprintf(
			"%s | %s | %.2f | %s",
			tx.CreatedAt.Format("2006-01-02"),
			tx.TypeTransaction,
			tx.Amount,
			tx.Category,
		))
		pdf.Ln(5)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return "", fmt.Errorf("failed to output PDF: %w", err)
	}

	key := fmt.Sprintf("exports/user_%d/export_%s.pdf", userID, time.Now().Format("2006-01-02"))
	if err := s.s3.Put(ctx, key, buf.Bytes()); err != nil {
		return "", fmt.Errorf("failed to save to S3: %w", err)
	}

	return key, nil
}
