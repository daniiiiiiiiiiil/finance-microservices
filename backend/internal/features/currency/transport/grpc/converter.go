package grpc

import (
	"backend/internal/core/domain"
	"backend/internal/features/currency/transport/grpc/proto"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func convertRateToProto(rate *domain.Rate) *proto.GetRatesResponse {
	return &proto.GetRatesResponse{
		Base:      rate.Base,
		Rates:     rate.Rates,
		Timestamp: timestamppb.New(rate.Timestamp),
	}
}

func convertConversionToProto(conversion *domain.Conversion) *proto.ConvertResponse {
	return &proto.ConvertResponse{
		From:      conversion.From,
		To:        conversion.To,
		Amount:    conversion.Amount,
		Result:    conversion.Result,
		Rate:      conversion.Rate,
		Timestamp: timestamppb.New(conversion.Timestamp),
	}
}

func convertProtoToConvertRequest(req *proto.ConvertRequest) *domain.Conversion {
	return &domain.Conversion{
		From:   req.From,
		To:     req.To,
		Amount: req.Amount,
	}
}
