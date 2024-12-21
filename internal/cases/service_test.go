package cases_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"project/internal/adapters/configs"
	"project/internal/cases"
	td "project/internal/cases/testdata"
	"project/internal/entities"
)

const (
	getAggregateMax     = "max"
	getAggregateMin     = "min"
	getAggregateAverage = "average"
	getAggregatePercent = "percent"
)

var (
	mockCrypProvError = errors.New("error processing request")
	mockStorageError  = errors.New("error processing data")
)

func Test_Service(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// "BTC", "ETH", "DOT", "ADA"

	var cfgService configs.Configs

	testCase := []struct {
		name            string
		ourActualTitles []string
		method          func(ctx context.Context, s *cases.Service) (interface{}, error)
		mockBehavior    func(ctx context.Context, mockCrProv *td.MockCryptoProvider, mockStrg *td.MockStorage)
		expectedValue   interface{}
		expectedError   error
	}{
		{
			name:            "GetLastRates - success. Where function checkExistTitles is not success",
			ourActualTitles: []string{"BTC"},
			method: func(ctx context.Context, s *cases.Service) (interface{}, error) {
				coin, err := s.GetLastRates(ctx, []string{"BTC", "ETH"})
				return coin, err
			},
			mockBehavior: func(ctx context.Context, mockCrProv *td.MockCryptoProvider, mockStrg *td.MockStorage) {
				mockTitles := []string{"BTC", "ETH"}
				mockCoinsNotExist := []entities.Coin{
					{
						Title:            "ETH",
						Cost:             46.6,
						RatePercentDelta: 0,
					},
				}

				mockCoins := []entities.Coin{
					{
						Title:            "BTC",
						Cost:             50.4,
						RatePercentDelta: 0,
					},
					{
						Title:            "ETH",
						Cost:             46.6,
						RatePercentDelta: 0,
					},
				}

				mockStrg.EXPECT().GetCoinsList(ctx).Return([]string{"BTC"}, nil).Times(1)
				mockCrProv.EXPECT().GetActualRates(ctx, []string{"ETH"}).Return(mockCoinsNotExist, nil).Times(1)
				mockStrg.EXPECT().Store(ctx, mockCoinsNotExist).Return(nil).Times(1)
				mockStrg.EXPECT().GetActualCoins(ctx, mockTitles).Return(mockCoins, nil).Times(1)
			},
			expectedValue: []entities.Coin{
				{
					Title:            "BTC",
					Cost:             50.4,
					RatePercentDelta: 0,
				},
				{
					Title:            "ETH",
					Cost:             46.6,
					RatePercentDelta: 0,
				},
			},
			expectedError: nil,
		},

		{
			name:            "GetLastRates - success. Where function checkExistTitles is success",
			ourActualTitles: []string{"BTC"},
			method: func(ctx context.Context, s *cases.Service) (interface{}, error) {
				coin, err := s.GetLastRates(ctx, []string{"BTC", "ETH", "ADA"})
				return coin, err
			},
			mockBehavior: func(ctx context.Context, mockCrProv *td.MockCryptoProvider, mockStrg *td.MockStorage) {
				mockexistTitles := []string{"BTC", "ETH", "DOT", "ADA"}
				mockTitles := []string{"BTC", "ETH", "ADA"}
				mockCoins := []entities.Coin{
					{
						Title:            "BTC",
						Cost:             50.4,
						RatePercentDelta: 0,
					},
					{
						Title:            "ETH",
						Cost:             46.6,
						RatePercentDelta: -4.775,
					},
					{
						Title:            "ADA",
						Cost:             81.2,
						RatePercentDelta: 0,
					},
				}

				mockStrg.EXPECT().GetCoinsList(ctx).Return(mockexistTitles, nil).Times(1)
				mockStrg.EXPECT().GetActualCoins(ctx, mockTitles).Return(mockCoins, nil).Times(1)
			},
			expectedValue: []entities.Coin{
				{
					Title:            "BTC",
					Cost:             50.4,
					RatePercentDelta: 0,
				},
				{
					Title:            "ETH",
					Cost:             46.6,
					RatePercentDelta: -4.775,
				},
				{
					Title:            "ADA",
					Cost:             81.2,
					RatePercentDelta: 0,
				},
			},
			expectedError: nil,
		},

		{
			name:            "GetLastRates - success. Where func GetCoinsList don`t have titles",
			ourActualTitles: []string{},
			method: func(ctx context.Context, s *cases.Service) (interface{}, error) {
				coin, err := s.GetLastRates(ctx, []string{"ADA"})
				return coin, err
			},
			mockBehavior: func(ctx context.Context, mockCrProv *td.MockCryptoProvider, mockStrg *td.MockStorage) {
				mockTitles := []string{"ADA"}
				mockCoins := []entities.Coin{
					{
						Title:            "ADA",
						Cost:             81.2,
						RatePercentDelta: 0,
					},
				}

				mockStrg.EXPECT().GetCoinsList(ctx).Return(nil, nil).Times(1)
				mockCrProv.EXPECT().GetActualRates(ctx, mockTitles).Return(mockCoins, nil).Times(1)
				mockStrg.EXPECT().Store(ctx, mockCoins).Return(nil).Times(1)
				mockStrg.EXPECT().GetActualCoins(ctx, mockTitles).Return(mockCoins, nil).Times(1)
			},
			expectedValue: []entities.Coin{
				{
					Title:            "ADA",
					Cost:             81.2,
					RatePercentDelta: 0,
				},
			},
			expectedError: nil,
		},

		{
			name:            "GetLastRates - not success. Where function GetActualRates return error",
			ourActualTitles: []string{},
			method: func(ctx context.Context, s *cases.Service) (interface{}, error) {
				coin, err := s.GetLastRates(ctx, []string{"ADA"})
				return coin, err
			},
			mockBehavior: func(ctx context.Context, mockCrProv *td.MockCryptoProvider, mockStrg *td.MockStorage) {
				mockTitles := []string{"ADA"}

				mockStrg.EXPECT().GetCoinsList(ctx).Return(nil, nil).Times(1)
				mockCrProv.EXPECT().GetActualRates(ctx, mockTitles).Return(nil, mockCrypProvError).Times(1)
			},
			expectedValue: []entities.Coin(nil),

			expectedError: mockCrypProvError,
		},

		{
			name:            "GetMaxReates -success",
			ourActualTitles: []string{"BTC", "ETH", "DOT", "ADA"},
			method: func(ctx context.Context, s *cases.Service) (interface{}, error) {
				coin, err := s.GetMaxRates(ctx, []string{"ETH"})
				return coin, err
			},
			mockBehavior: func(ctx context.Context, mockCrProv *td.MockCryptoProvider, mockStrg *td.MockStorage) {
				mockRequestTitles := []string{"ETH"}
				mockTitles := []string{"BTC", "ETH", "DOT", "ADA"}
				mockCoins := []entities.Coin{
					{
						Title:            "ETH",
						Cost:             90.6,
						RatePercentDelta: 2.13,
					},
				}

				mockStrg.EXPECT().GetCoinsList(ctx).Return(mockTitles, nil).Times(1)
				mockStrg.EXPECT().GetAggregateCoins(ctx, mockRequestTitles, getAggregateMax).Return(mockCoins, nil).Times(1)
			},
			expectedValue: []entities.Coin{
				{
					Title:            "ETH",
					Cost:             90.6,
					RatePercentDelta: 2.13,
				},
			},
			expectedError: nil,
		},

		{
			name:            "GetMaxReates -not success. Where func GetAggregateCoins returned error", //
			ourActualTitles: []string{"ETH", "DOT", "ADA"},
			method: func(ctx context.Context, s *cases.Service) (interface{}, error) {
				coin, err := s.GetMaxRates(ctx, []string{"DOT"})
				return coin, err
			},
			mockBehavior: func(ctx context.Context, mockCrProv *td.MockCryptoProvider, mockStrg *td.MockStorage) {
				mockTitles := []string{"ETH", "DOT", "ADA"}

				mockStrg.EXPECT().GetCoinsList(ctx).Return(mockTitles, nil).Times(1)
				mockStrg.EXPECT().GetAggregateCoins(ctx, []string{"DOT"}, getAggregateMax).Return(nil, mockStorageError).Times(1)
			},
			expectedValue: []entities.Coin(nil),
			expectedError: mockStorageError,
		},

		{
			name:            "GetMaxReates -not success. Where GetCoinsList returned error",
			ourActualTitles: []string{"BTC", "ETH", "DOT", "ADA"},
			method: func(ctx context.Context, s *cases.Service) (interface{}, error) {
				coin, err := s.GetMaxRates(ctx, []string{"ETH"})
				return coin, err
			},
			mockBehavior: func(ctx context.Context, mockCrProv *td.MockCryptoProvider, mockStrg *td.MockStorage) {
				mockStrg.EXPECT().GetCoinsList(ctx).Return(nil, mockStorageError).Times(1)
			},
			expectedValue: []entities.Coin(nil),
			expectedError: mockStorageError,
		},

		{
			name:            "GetMinReates -success",
			ourActualTitles: []string{"BTC", "ETH", "DOT", "ADA"},
			method: func(ctx context.Context, s *cases.Service) (interface{}, error) {
				coin, err := s.GetMinRates(ctx, []string{"ETH"})
				return coin, err
			},
			mockBehavior: func(ctx context.Context, mockCrProv *td.MockCryptoProvider, mockStrg *td.MockStorage) {
				mockRequestTitles := []string{"ETH"}
				mockTitles := []string{"BTC", "ETH", "DOT", "ADA"}
				mockCoins := []entities.Coin{
					{
						Title:            "ETH",
						Cost:             90.6,
						RatePercentDelta: 2.13,
					},
				}

				mockStrg.EXPECT().GetCoinsList(ctx).Return(mockTitles, nil).Times(1)
				mockStrg.EXPECT().GetAggregateCoins(ctx, mockRequestTitles, getAggregateMin).Return(mockCoins, nil).Times(1)
			},
			expectedValue: []entities.Coin{
				{
					Title:            "ETH",
					Cost:             90.6,
					RatePercentDelta: 2.13,
				},
			},
			expectedError: nil,
		},

		{
			name:            "GetMinReates -success",
			ourActualTitles: []string{"ETH", "DOT", "ADA"},
			method: func(ctx context.Context, s *cases.Service) (interface{}, error) {
				coin, err := s.GetMinRates(ctx, []string{"ETH"})
				return coin, err
			},
			mockBehavior: func(ctx context.Context, mockCrProv *td.MockCryptoProvider, mockStrg *td.MockStorage) {
				mockTitles := []string{"ETH", "DOT", "ADA"}
				mockCoins := []entities.Coin{
					{
						Title:            "ETH",
						Cost:             90.6,
						RatePercentDelta: 2.13,
					},
				}

				mockStrg.EXPECT().GetCoinsList(ctx).Return(mockTitles, nil).Times(1)
				mockStrg.EXPECT().GetAggregateCoins(ctx, []string{"ETH"}, getAggregateMin).Return(mockCoins, nil).Times(1)
			},
			expectedValue: []entities.Coin{
				{
					Title:            "ETH",
					Cost:             90.6,
					RatePercentDelta: 2.13,
				},
			},
			expectedError: nil,
		},

		{
			name:            "GetMinReates -not success. Where GetAggregateCoins returned error",
			ourActualTitles: []string{"BTC", "ETH", "DOT", "ADA"},
			method: func(ctx context.Context, s *cases.Service) (interface{}, error) {
				coin, err := s.GetMinRates(ctx, []string{"ETH"})
				return coin, err
			},
			mockBehavior: func(ctx context.Context, mockCrProv *td.MockCryptoProvider, mockStrg *td.MockStorage) {
				mockRequestTitles := []string{"ETH"}
				mockTitles := []string{"BTC", "ETH", "DOT", "ADA"}

				mockStrg.EXPECT().GetCoinsList(ctx).Return(mockTitles, nil).Times(1)
				mockStrg.EXPECT().GetAggregateCoins(ctx, mockRequestTitles, getAggregateMin).Return(nil, mockStorageError).Times(1)
			},
			expectedValue: []entities.Coin(nil),
			expectedError: mockStorageError,
		},

		{
			name:            "GetAvgReates -success",
			ourActualTitles: []string{"BTC", "ETH", "DOT", "ADA"},
			method: func(ctx context.Context, s *cases.Service) (interface{}, error) {
				coin, err := s.GetAvgRates(ctx, []string{"ETH", "ADA"})
				return coin, err
			},
			mockBehavior: func(ctx context.Context, mockCrProv *td.MockCryptoProvider, mockStrg *td.MockStorage) {
				mockRequestTitles := []string{"ETH", "ADA"}
				mockTitles := []string{"BTC", "ETH", "DOT", "ADA"}
				mockCoins := []entities.Coin{
					{
						Title:            "ETH",
						Cost:             90.6,
						RatePercentDelta: 2.13,
					},
					{
						Title:            "ADA",
						Cost:             77.2,
						RatePercentDelta: -1.35,
					},
				}

				mockStrg.EXPECT().GetCoinsList(ctx).Return(mockTitles, nil).Times(1)
				mockStrg.EXPECT().GetAggregateCoins(ctx, mockRequestTitles, getAggregateAverage).Return(mockCoins, nil).Times(1)
			},
			expectedValue: []entities.Coin{
				{
					Title:            "ETH",
					Cost:             90.6,
					RatePercentDelta: 2.13,
				},
				{
					Title:            "ADA",
					Cost:             77.2,
					RatePercentDelta: -1.35,
				},
			},
			expectedError: nil,
		},

		{
			name:            "GetAvgReates -not success. Where func GetCoinsList returned error",
			ourActualTitles: []string{"ETH", "DOT", "ADA"},
			method: func(ctx context.Context, s *cases.Service) (interface{}, error) {
				coin, err := s.GetAvgRates(ctx, []string{"BTC"})
				return coin, err
			},
			mockBehavior: func(ctx context.Context, mockCrProv *td.MockCryptoProvider, mockStrg *td.MockStorage) {

				mockStrg.EXPECT().GetCoinsList(ctx).Return(nil, mockStorageError).Times(1)
			},
			expectedValue: []entities.Coin(nil),
			expectedError: mockStorageError,
		},

		{
			name:            "GetAvgReates -not success. Where GetAggregateCoins returned error",
			ourActualTitles: []string{"BTC", "ETH", "DOT", "ADA"},
			method: func(ctx context.Context, s *cases.Service) (interface{}, error) {
				coin, err := s.GetAvgRates(ctx, []string{"ETH"})
				return coin, err
			},
			mockBehavior: func(ctx context.Context, mockCrProv *td.MockCryptoProvider, mockStrg *td.MockStorage) {
				mockRequestTitles := []string{"ETH"}
				mockTitles := []string{"BTC", "ETH", "DOT", "ADA"}

				mockStrg.EXPECT().GetCoinsList(ctx).Return(mockTitles, nil).Times(1)
				mockStrg.EXPECT().GetAggregateCoins(ctx, mockRequestTitles, getAggregateAverage).Return(nil, mockStorageError).Times(1)
			},
			expectedValue: []entities.Coin(nil),
			expectedError: mockStorageError,
		},

		{
			name:            "GetPercentReates -success",
			ourActualTitles: []string{"ETH", "DOT", "ADA"},
			method: func(ctx context.Context, s *cases.Service) (interface{}, error) {
				coin, err := s.GetPercentRates(ctx, []string{"ETH"})
				return coin, err
			},
			mockBehavior: func(ctx context.Context, mockCrProv *td.MockCryptoProvider, mockStrg *td.MockStorage) {
				mockTitles := []string{"ETH", "DOT", "ADA"}
				mockCoins := []entities.Coin{
					{
						Title:            "ETH",
						Cost:             90.6,
						RatePercentDelta: 2.13,
					},
				}

				mockStrg.EXPECT().GetCoinsList(ctx).Return(mockTitles, nil).Times(1)
				mockStrg.EXPECT().GetAggregateCoins(ctx, []string{"ETH"}, getAggregatePercent).Return(mockCoins, nil).Times(1)
			},
			expectedValue: []entities.Coin{
				{
					Title:            "ETH",
					Cost:             90.6,
					RatePercentDelta: 2.13,
				},
			},
			expectedError: nil,
		},

		{
			name:            "GetPercentReates -not success. Where func GetCoinsList returned error",
			ourActualTitles: []string{"ETH", "DOT", "ADA"},
			method: func(ctx context.Context, s *cases.Service) (interface{}, error) {
				coin, err := s.GetPercentRates(ctx, []string{"BTC"})
				return coin, err
			},
			mockBehavior: func(ctx context.Context, mockCrProv *td.MockCryptoProvider, mockStrg *td.MockStorage) {

				mockStrg.EXPECT().GetCoinsList(ctx).Return(nil, mockStorageError).Times(1)
			},
			expectedValue: []entities.Coin(nil),
			expectedError: mockStorageError,
		},

		{
			name:            "GetPercentReates -not success. Where GetAggregateCoins returned error",
			ourActualTitles: []string{"BTC", "ETH", "DOT", "ADA"},
			method: func(ctx context.Context, s *cases.Service) (interface{}, error) {
				coin, err := s.GetPercentRates(ctx, []string{"ETH"})
				return coin, err
			},
			mockBehavior: func(ctx context.Context, mockCrProv *td.MockCryptoProvider, mockStrg *td.MockStorage) {
				mockRequestTitles := []string{"ETH"}
				mockTitles := []string{"BTC", "ETH", "DOT", "ADA"}

				mockStrg.EXPECT().GetCoinsList(ctx).Return(mockTitles, nil).Times(1)
				mockStrg.EXPECT().GetAggregateCoins(ctx, mockRequestTitles, getAggregatePercent).Return(nil, mockStorageError).Times(1)
			},
			expectedValue: []entities.Coin(nil),
			expectedError: mockStorageError,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			mockCryptProv := td.NewMockCryptoProvider(mockCtrl)
			mockStorage := td.NewMockStorage(mockCtrl)

			ctx := context.Background()

			tc.mockBehavior(ctx, mockCryptProv, mockStorage)

			cfgService.ActualTitles = tc.ourActualTitles
			service, err := cases.NewService(mockCryptProv, mockStorage, &cfgService)
			require.ErrorIs(t, err, nil)

			value, err := tc.method(ctx, service)

			require.Equal(t, tc.expectedValue, value)
			require.ErrorIs(t, err, tc.expectedError)

		})
	}

}
