package service_currency

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/currency-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/currency-service/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockRateCache struct {
	mock.Mock
}

func (m *MockRateCache) GetRate(ctx context.Context, base string) (*domain.Rate, error) {
	args := m.Called(ctx, base)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Rate), args.Error(1)
}

func (m *MockRateCache) SetRate(ctx context.Context, rate domain.Rate, ttl time.Duration) error {
	args := m.Called(ctx, rate, ttl)
	return args.Error(0)
}

func (m *MockRateCache) GetConvertedUSD(ctx context.Context, txID int) (float64, error) {
	args := m.Called(ctx, txID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockRateCache) SetConvertedUSD(ctx context.Context, txID int, amount float64, ttl time.Duration) error {
	args := m.Called(ctx, txID, amount, ttl)
	return args.Error(0)
}

func (m *MockRateCache) DeleteRate(ctx context.Context, base string) error {
	args := m.Called(ctx, base)
	return args.Error(0)
}

func (m *MockRateCache) DeleteConvertedUSD(ctx context.Context, txID int) error {
	args := m.Called(ctx, txID)
	return args.Error(0)
}

func (m *MockRateCache) Exists(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRateCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockRateCache) SetTransactionUSD(ctx context.Context, tx domain.TransactionUSD, ttl time.Duration) error {
	args := m.Called(ctx, tx, ttl)
	return args.Error(0)
}

func (m *MockRateCache) GetTransactionUSD(ctx context.Context, txID int) (domain.TransactionUSD, error) {
	args := m.Called(ctx, txID)
	return args.Get(0).(domain.TransactionUSD), args.Error(1)
}

type MockCurrencyAPIClient struct {
	mock.Mock
}

func (m *MockCurrencyAPIClient) GetRatesFromAPI(ctx context.Context, base string) (*domain.Rate, error) {
	args := m.Called(ctx, base)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Rate), args.Error(1)
}

type MockRedis struct {
	mock.Mock
}

func (m *MockRedis) Get(ctx context.Context, key string, dest interface{}) error {
	args := m.Called(ctx, key, dest)
	return args.Error(0)
}

func (m *MockRedis) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockRedis) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockRedis) Incr(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRedis) Expire(ctx context.Context, key string, ttl time.Duration) error {
	args := m.Called(ctx, key, ttl)
	return args.Error(0)
}

func (m *MockRedis) Exists(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) Publish(ctx context.Context, eventType string, data interface{}) error {
	args := m.Called(ctx, eventType, data)
	return args.Error(0)
}

type testSuite struct {
	service       *CurrencyService
	mockRateCache *MockRateCache
	mockClient    *MockCurrencyAPIClient
	mockRedis     *MockRedis
	mockPublisher *MockPublisher
}

func setup() *testSuite {
	mockRateCache := new(MockRateCache)
	mockClient := new(MockCurrencyAPIClient)
	mockRedis := new(MockRedis)
	mockPublisher := new(MockPublisher)

	loggerInstance := &logger.Logger{Logger: zap.NewNop()}

	service := &CurrencyService{
		rateCache:      mockRateCache,
		client:         mockClient,
		logger:         loggerInstance,
		redis:          mockRedis,
		eventPublisher: mockPublisher,
	}

	return &testSuite{
		service:       service,
		mockRateCache: mockRateCache,
		mockClient:    mockClient,
		mockRedis:     mockRedis,
		mockPublisher: mockPublisher,
	}
}

func TestConvert_Success(t *testing.T) {
	s := setup()
	ctx := context.Background()

	convertReq := &domain.Conversion{
		From:   "USD",
		To:     "EUR",
		Amount: 100,
	}

	rate := &domain.Rate{
		Base: "USD",
		Rates: map[string]float64{
			"EUR": 0.85,
			"RUB": 91.5,
		},
	}

	s.mockRedis.On("Get", ctx, "currency:rate:USD", mock.Anything).Return(errors.New("not found"))
	s.mockRateCache.On("GetRate", ctx, "USD").Return(rate, nil)
	s.mockRateCache.On("SetRate", ctx, *rate, time.Hour).Return(nil)

	result, err := s.service.Convert(ctx, convertReq)

	assert.NoError(t, err)
	assert.Equal(t, 85.0, result.Result)
	assert.Equal(t, 0.85, result.Rate)
	assert.Equal(t, "USD", result.From)
	assert.Equal(t, "EUR", result.To)
}

func TestConvert_ValidationError(t *testing.T) {
	s := setup()
	ctx := context.Background()

	convertReq := &domain.Conversion{
		From:   "",
		To:     "EUR",
		Amount: 100,
	}

	_, err := s.service.Convert(ctx, convertReq)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid conversion")
}

func TestConvert_CurrencyNotFound(t *testing.T) {
	s := setup()
	ctx := context.Background()

	convertReq := &domain.Conversion{
		From:   "USD",
		To:     "XYZ",
		Amount: 100,
	}

	rate := &domain.Rate{
		Base: "USD",
		Rates: map[string]float64{
			"EUR": 0.85,
		},
	}

	s.mockRedis.On("Get", ctx, "currency:rate:USD", mock.Anything).Return(errors.New("not found"))
	s.mockRateCache.On("GetRate", ctx, "USD").Return(rate, nil)
	s.mockRateCache.On("SetRate", ctx, *rate, time.Hour).Return(nil)

	_, err := s.service.Convert(ctx, convertReq)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "currency XYZ not found")
}

func TestConvertBatch_Success(t *testing.T) {
	s := setup()
	ctx := context.Background()

	rate := &domain.Rate{
		Base: "USD",
		Rates: map[string]float64{
			"EUR": 0.85,
			"RUB": 91.5,
		},
	}

	s.mockRedis.On("Get", ctx, "currency:rate:USD", mock.Anything).Return(errors.New("not found"))
	s.mockRateCache.On("GetRate", ctx, "USD").Return(rate, nil)
	s.mockRateCache.On("SetRate", ctx, *rate, time.Hour).Return(nil)

	result, err := s.service.ConvertBatch(ctx, "USD", []string{"EUR", "RUB"}, 100)

	assert.NoError(t, err)
	assert.Equal(t, 85.0, result["EUR"])
	assert.Equal(t, 9150.0, result["RUB"])
}

func TestConvertBatch_InvalidAmount(t *testing.T) {
	s := setup()
	ctx := context.Background()

	_, err := s.service.ConvertBatch(ctx, "USD", []string{"EUR"}, -100)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "amount must be greater than zero")
}

func TestGetRates_FromCache(t *testing.T) {
	s := setup()
	ctx := context.Background()

	rate := &domain.Rate{
		Base: "USD",
		Rates: map[string]float64{
			"EUR": 0.85,
		},
		Timestamp: time.Now(),
	}

	s.mockRedis.On("Get", ctx, "currency:rate:USD", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*domain.Rate)
		*dest = *rate
	})

	result, err := s.service.GetRates(ctx, "USD")

	assert.NoError(t, err)
	assert.Equal(t, rate.Base, result.Base)
	assert.Equal(t, rate.Rates, result.Rates)
}

func TestGetRates_FromAPI(t *testing.T) {
	s := setup()
	ctx := context.Background()

	rateFromAPI := &domain.Rate{
		Base: "USD",
		Rates: map[string]float64{
			"EUR": 0.85,
		},
		Timestamp: time.Now(),
	}

	s.mockRedis.On("Get", ctx, "currency:rate:USD", mock.Anything).Return(errors.New("not found"))

	s.mockRateCache.On("GetRate", ctx, "USD").Return(nil, errors.New("not found"))
	s.mockClient.On("GetRatesFromAPI", ctx, "USD").Return(rateFromAPI, nil)

	s.mockRateCache.On("SetRate", ctx, *rateFromAPI, time.Hour).Return(nil)

	result, err := s.service.GetRates(ctx, "USD")

	assert.NoError(t, err)
	assert.Equal(t, rateFromAPI, result)
}

func TestFetchRates_FromCache(t *testing.T) {
	s := setup()
	ctx := context.Background()

	rate := &domain.Rate{
		Base: "USD",
		Rates: map[string]float64{
			"EUR": 0.85,
		},
		Timestamp: time.Now(),
	}

	s.mockRateCache.On("GetRate", ctx, "USD").Return(rate, nil)

	result, err := s.service.FetchRates(ctx, "USD")

	assert.NoError(t, err)
	assert.Equal(t, rate, result)
}

func TestFetchRates_FromAPI(t *testing.T) {
	s := setup()
	ctx := context.Background()

	rateFromAPI := &domain.Rate{
		Base: "USD",
		Rates: map[string]float64{
			"EUR": 0.85,
		},
		Timestamp: time.Now(),
	}

	s.mockRateCache.On("GetRate", ctx, "USD").Return(nil, errors.New("not found"))
	s.mockClient.On("GetRatesFromAPI", ctx, "USD").Return(rateFromAPI, nil)
	s.mockRateCache.On("SetRate", ctx, *rateFromAPI, time.Hour).Return(nil)

	result, err := s.service.FetchRates(ctx, "USD")

	assert.NoError(t, err)
	assert.Equal(t, rateFromAPI, result)
}

func TestGetTransactionUSD_Success(t *testing.T) {
	s := setup()
	ctx := context.Background()

	tx := domain.TransactionUSD{
		TransactionID:    1,
		AmountUSD:        85.0,
		OriginalAmount:   100,
		OriginalCurrency: "EUR",
		ConvertedAt:      time.Now(),
	}

	s.mockRateCache.On("GetTransactionUSD", ctx, 1).Return(tx, nil)

	result, err := s.service.GetTransactionUSD(ctx, 1)

	assert.NoError(t, err)
	assert.Equal(t, tx, result)
}

func TestGetTransactionUSD_InvalidID(t *testing.T) {
	s := setup()
	ctx := context.Background()

	_, err := s.service.GetTransactionUSD(ctx, 0)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "txID must be greater than zero")
}

func TestGetTransactionUSD_NotFound(t *testing.T) {
	s := setup()
	ctx := context.Background()

	s.mockRateCache.On("GetTransactionUSD", ctx, 1).Return(domain.TransactionUSD{}, errors.New("not found"))

	_, err := s.service.GetTransactionUSD(ctx, 1)

	assert.Error(t, err)
}

func TestConvert_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		from        string
		to          string
		amount      float64
		rates       map[string]float64
		expectError bool
		expected    float64
	}{
		{
			name:     "USD to EUR",
			from:     "USD",
			to:       "EUR",
			amount:   100,
			rates:    map[string]float64{"EUR": 0.85},
			expected: 85.0,
		},
		{
			name:     "USD to RUB",
			from:     "USD",
			to:       "RUB",
			amount:   100,
			rates:    map[string]float64{"RUB": 91.5},
			expected: 9150.0,
		},
		{
			name:        "currency not found",
			from:        "USD",
			to:          "XYZ",
			amount:      100,
			rates:       map[string]float64{"EUR": 0.85},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := setup()
			ctx := context.Background()

			convertReq := &domain.Conversion{
				From:   tt.from,
				To:     tt.to,
				Amount: tt.amount,
			}

			rate := &domain.Rate{
				Base:  tt.from,
				Rates: tt.rates,
			}

			s.mockRedis.On("Get", ctx, "currency:rate:"+tt.from, mock.Anything).Return(errors.New("not found"))
			s.mockRateCache.On("GetRate", ctx, tt.from).Return(rate, nil)
			s.mockRateCache.On("SetRate", ctx, *rate, time.Hour).Return(nil)

			result, err := s.service.Convert(ctx, convertReq)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result.Result)
			}
		})
	}
}

func TestConvertBatch_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		from        string
		toList      []string
		amount      float64
		rates       map[string]float64
		expectError bool
		expected    map[string]float64
	}{
		{
			name:     "multiple currencies",
			from:     "USD",
			toList:   []string{"EUR", "RUB"},
			amount:   100,
			rates:    map[string]float64{"EUR": 0.85, "RUB": 91.5},
			expected: map[string]float64{"EUR": 85.0, "RUB": 9150.0},
		},
		{
			name:        "invalid amount",
			from:        "USD",
			toList:      []string{"EUR"},
			amount:      -100,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := setup()
			ctx := context.Background()

			if tt.amount > 0 {
				rate := &domain.Rate{
					Base:  tt.from,
					Rates: tt.rates,
				}

				s.mockRedis.On("Get", ctx, "currency:rate:"+tt.from, mock.Anything).Return(errors.New("not found"))
				s.mockRateCache.On("GetRate", ctx, tt.from).Return(rate, nil)
				s.mockRateCache.On("SetRate", ctx, *rate, time.Hour).Return(nil)
			}

			result, err := s.service.ConvertBatch(ctx, tt.from, tt.toList, tt.amount)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
