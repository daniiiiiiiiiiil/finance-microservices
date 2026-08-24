package grpc

import (
	"backend/internal/core/domain"
	"backend/proto/currency/gen"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func convertRateToProto(rate *domain.Rate) *gen.GetRatesResponse {
	return &gen.GetRatesResponse{
		Base:      rate.Base,
		Rates:     rate.Rates,
		Timestamp: timestamppb.New(rate.Timestamp),
	}
}

func convertConversionToProto(conversion *domain.Conversion) *gen.ConvertResponse {
	return &gen.ConvertResponse{
		From:      conversion.From,
		To:        conversion.To,
		Amount:    conversion.Amount,
		Result:    conversion.Result,
		Rate:      conversion.Rate,
		Timestamp: timestamppb.New(conversion.Timestamp),
	}
}

func convertProtoToConvertRequest(req *gen.ConvertRequest) *domain.Conversion {
	return &domain.Conversion{
		From:   req.From,
		To:     req.To,
		Amount: req.Amount,
	}
}
