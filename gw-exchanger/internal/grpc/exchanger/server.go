package exchanger

import (
	"context"
	"gw-exchanger/internal/domain/models"
	"log/slog"
	"strings"

	exchangev1 "github.com/yaoguao/gw/protos/gen/go/exchange"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ExchangeRatesGetter interface {
	GetExchangeRates(ctx context.Context, base string) (*models.ExchangeRatesResponse, error)
	GetExchangeRateForCurrency(ctx context.Context, from, to string) (*models.ExchangeRates, error)
}

type serverAPI struct {
	exchangev1.UnimplementedExchangeServiceServer

	erGetter ExchangeRatesGetter

	log *slog.Logger
}

func Register(gRPC *grpc.Server, erGetter ExchangeRatesGetter, log *slog.Logger) {
	exchangev1.RegisterExchangeServiceServer(gRPC, &serverAPI{
		erGetter: erGetter,
		log:      log,
	})
}

func (s serverAPI) GetExchangeRates(ctx context.Context, request *exchangev1.ExchangeRateRequest) (*exchangev1.ExchangeRatesResponse, error) {
	if err := validExchangeRate(request); err != nil {
		s.log.Debug("err", err)
		return nil, err
	}

	exResponse, err := s.erGetter.GetExchangeRates(ctx, request.GetBase())
	if err != nil {
		s.log.Debug("err", err)

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &exchangev1.ExchangeRatesResponse{
		Rates: exResponse.Rates,
	}, nil
}

func (s serverAPI) GetExchangeRateForCurrency(ctx context.Context, request *exchangev1.CurrencyRequest) (*exchangev1.ExchangeRateResponse, error) {
	if err := validCurrencyRequest(request); err != nil {
		s.log.Debug("err", err)
		return nil, err
	}
	exchangeRate, err := s.erGetter.GetExchangeRateForCurrency(ctx, request.GetFromCurrency(), request.GetToCurrency())
	if err != nil {
		s.log.Debug("err", err.Error())

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &exchangev1.ExchangeRateResponse{
		FromCurrency: exchangeRate.FromCurrency,
		ToCurrency:   exchangeRate.ToCurrency,
		Rate:         exchangeRate.Rate,
	}, nil
}

func validExchangeRate(req *exchangev1.ExchangeRateRequest) error {
	if len(req.GetBase()) != 3 {
		return status.Error(codes.InvalidArgument, "name base currency does not the format")
	}

	return nil
}

func validCurrencyRequest(req *exchangev1.CurrencyRequest) error {
	if len(req.GetToCurrency()) != 3 {
		return status.Error(codes.InvalidArgument, "name FromCurrency currency does not the format")
	}

	if len(req.GetToCurrency()) != 3 {
		return status.Error(codes.InvalidArgument, "name ToCurrency currency does not the format")
	}

	if strings.EqualFold(req.GetToCurrency(), req.GetFromCurrency()) {
		return status.Error(codes.InvalidArgument, "you cannot get the exchange rate of the same currency")
	}

	return nil
}
