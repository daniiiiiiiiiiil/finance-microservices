package grpc

import (
	"context"

	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/pkg/grpcutil/interceptors"
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/proto/finance/gen"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *FinanceServer) ExportJSON(ctx context.Context, req *gen.ExportRequest) (*gen.ExportResponse, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	key, err := s.service.ExportJSON(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to export JSON")
	}
	return &gen.ExportResponse{
		Key: key,
	}, nil
}

func (s *FinanceServer) ExportCSV(ctx context.Context, req *gen.ExportRequest) (*gen.ExportResponse, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	key, err := s.service.ExportCSV(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to export CSV")
	}
	return &gen.ExportResponse{
		Key: key,
	}, nil
}

func (s *FinanceServer) ExportTXT(ctx context.Context, req *gen.ExportRequest) (*gen.ExportResponse, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	key, err := s.service.ExportTXT(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to export TXT")
	}
	return &gen.ExportResponse{
		Key: key,
	}, nil
}

func (s *FinanceServer) ExportPDF(ctx context.Context, req *gen.ExportRequest) (*gen.ExportResponse, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	key, err := s.service.ExportPDF(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to export PDF")
	}

	return &gen.ExportResponse{
		Key: key,
	}, nil
}

func (s *FinanceServer) DownloadExport(ctx context.Context, req *gen.DownloadExportRequest) (*gen.DownloadExportResponse, error) {
	_, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	data, err := s.service.DownloadExport(ctx, req.Key)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to download export")
	}

	return &gen.DownloadExportResponse{
		Data: data,
	}, nil
}
