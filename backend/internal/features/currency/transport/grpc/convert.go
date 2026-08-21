package grpc

import (
	"backend/internal/core/transport/grpc/interceptors"
	"backend/internal/features/currency/transport/grpc/proto"
	"time"

	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *CurrencyServer) Convert(ctx context.Context, req *proto.ConvertRequest) (*proto.ConvertResponse, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	s.logger.Debug("gRPC Convert",
		zap.Int("user_id", userID),
		zap.String("from", req.From),
		zap.String("to", req.To),
		zap.Float64("amount", req.Amount),
	)

	conversion := convertProtoToConvertRequest(req)

	convert, err := s.service.Convert(ctx, conversion)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return convertConversionToProto(convert), nil
}

func (s *CurrencyServer) ConvertBatch(ctx context.Context, req *proto.ConvertBatchRequest) (*proto.ConvertBatchResponse, error) {
	UserID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "User ID is not found")
	}
	s.logger.Debug("gRPC ConvertBatch", zap.Int("user_id", UserID))
	result, err := s.service.ConvertBatch(ctx, req.From, req.To, req.Amount)
	if err != nil {
		s.logger.Error("gRPC ConvertBatch error", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &proto.ConvertBatchResponse{
		From:      req.From,
		Results:   result,
		Amount:    req.Amount,
		Timestamp: timestamppb.New(time.Now()),
	}, nil
}
