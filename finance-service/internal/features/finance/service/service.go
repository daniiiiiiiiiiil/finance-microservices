package service

import (
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/ports"
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/repository/postgres/pool"
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/pkg/logger"
	"golang.org/x/net/context"
)

var _ ports.FinanceServiceInterface = (*FinanceService)(nil)

type FinanceService struct {
	repo           ports.FinanceRepositoryInterface
	pool           pool.Pool
	redis          ports.RedisInterface
	eventPublisher ports.EventPublisherInterface
	exportService  *ExportService
	logger         *logger.Logger
}

func NewFinanceService(
	repo ports.FinanceRepositoryInterface,
	pool pool.Pool,
	redis ports.RedisInterface,
	eventPublisher ports.EventPublisherInterface,
	exportService *ExportService,
	logger *logger.Logger,
) *FinanceService {
	return &FinanceService{
		repo:           repo,
		pool:           pool,
		redis:          redis,
		eventPublisher: eventPublisher,
		exportService:  exportService,
		logger:         logger,
	}
}

func (s *FinanceService) ExportJSON(ctx context.Context, userID int) (string, error) {
	return s.exportService.ExportJSON(ctx, userID)
}

func (s *FinanceService) ExportCSV(ctx context.Context, userID int) (string, error) {
	return s.exportService.ExportCSV(ctx, userID)
}

func (s *FinanceService) ExportTXT(ctx context.Context, userID int) (string, error) {
	return s.exportService.ExportTXT(ctx, userID)
}

func (s *FinanceService) ExportPDF(ctx context.Context, userID int) (string, error) {
	return s.exportService.ExportPDF(ctx, userID)
}

func (s *FinanceService) DownloadExport(ctx context.Context, key string) ([]byte, error) {
	return s.exportService.s3.Get(ctx, key)
}

func (s *FinanceService) GetExportURL(ctx context.Context, key string) (string, error) {
	return s.exportService.GetExportURL(ctx, key)
}
