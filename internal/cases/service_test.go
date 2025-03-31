package cases_test

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"crypto-project/internal/cases"
	"crypto-project/internal/cases/mocks"
	"crypto-project/internal/entities"
)

var (
	ErrTest = errors.New("test error")
)

func TestGetAggregateRates(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockCryptoProvider := mocks.NewMockCryptoProvider(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	service, _ := cases.NewService(mockCryptoProvider, mockStorage, mockLogger)

	testTable := []struct {
		name        string
		titles      []string
		setupMock   func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider, mockLogger *mocks.MockLogger)
		expectedRes []*entities.Coin
		wantErr     bool
		expectedErr error
	}{
		{
			name:   "valid params, all coins stored",
			titles: []string{"Bitcoin", "ETC", "TON"},
			setupMock: func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider, mockLogger *mocks.MockLogger) {
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return([]string{"Bitcoin", "ETC", "TON"}, nil)
				mockLogger.EXPECT().
					Warn(gomock.Any()).
					Return()
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return([]string{"Bitcoin", "ETC", "TON"}, nil)
				mockStorage.EXPECT().
					GetAggregateCoins(gomock.Any(), []string{"Bitcoin", "ETC", "TON"}, gomock.Any()).
					Return([]*entities.Coin{
						{Title: "Bitcoin", Cost: 1000},
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}, nil)
			},
			expectedRes: []*entities.Coin{
				{Title: "Bitcoin", Cost: 1000},
				{Title: "ETH", Cost: 5555},
				{Title: "TON", Cost: 1},
			},
			wantErr: false,
		},
		{
			name:   "valid params, not all coins stored",
			titles: []string{"Bitcoin", "ETC", "TON"},
			setupMock: func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider, mockLogger *mocks.MockLogger) {
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return([]string{"Bitcoin", "ETC", "TON"}, nil)
				mockCryptoProvider.EXPECT().
					GetActualRates(gomock.Any(), []string{"Bitcoin", "ETC", "TON"}).
					Return([]*entities.Coin{
						{Title: "Bitcoin", Cost: 1000},
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}, nil)
				mockStorage.EXPECT().
					Store(gomock.Any(), []*entities.Coin{
						{Title: "Bitcoin", Cost: 1000},
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}).
					Return(nil)
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return([]string{"Bitcoin"}, nil)
				mockCryptoProvider.EXPECT().
					GetActualRates(gomock.Any(), []string{"ETC", "TON"}).
					Return([]*entities.Coin{
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}, nil)
				mockStorage.EXPECT().
					Store(gomock.Any(), []*entities.Coin{
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}).
					Return(nil)
				mockStorage.EXPECT().
					GetAggregateCoins(gomock.Any(), []string{"Bitcoin", "ETC", "TON"}, gomock.Any()).
					Return([]*entities.Coin{
						{Title: "Bitcoin", Cost: 1000},
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}, nil)
			},
			expectedRes: []*entities.Coin{
				{Title: "Bitcoin", Cost: 1000},
				{Title: "ETH", Cost: 5555},
				{Title: "TON", Cost: 1},
			},
			wantErr: false,
		},
		{
			name:   "valid params, error processNotExistTitles(GetCoinsList)",
			titles: []string{"Bitcoin", "ETC", "TON"},
			setupMock: func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider, mockLogger *mocks.MockLogger) {
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return([]string{"Bitcoin", "ETC", "TON"}, nil)
				mockCryptoProvider.EXPECT().
					GetActualRates(gomock.Any(), []string{"Bitcoin", "ETC", "TON"}).
					Return([]*entities.Coin{
						{Title: "Bitcoin", Cost: 1000},
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}, nil)
				mockStorage.EXPECT().
					Store(gomock.Any(), []*entities.Coin{
						{Title: "Bitcoin", Cost: 1000},
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}).
					Return(nil)
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return(nil, ErrTest)
			},
			expectedRes: nil,
			wantErr:     true,
			expectedErr: ErrTest,
		},
		{
			name:   "valid params, error processNotExistTitles(GetActualRates)",
			titles: []string{"Bitcoin", "ETC", "TON"},
			setupMock: func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider, mockLogger *mocks.MockLogger) {
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return([]string{"Bitcoin", "ETC", "TON"}, nil)
				mockCryptoProvider.EXPECT().
					GetActualRates(gomock.Any(), []string{"Bitcoin", "ETC", "TON"}).
					Return([]*entities.Coin{
						{Title: "Bitcoin", Cost: 1000},
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}, nil)
				mockStorage.EXPECT().
					Store(gomock.Any(), []*entities.Coin{
						{Title: "Bitcoin", Cost: 1000},
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}).
					Return(nil)
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return([]string{"Bitcoin"}, nil)
				mockCryptoProvider.EXPECT().
					GetActualRates(gomock.Any(), []string{"ETC", "TON"}).
					Return(nil, ErrTest)
			},
			expectedRes: nil,
			wantErr:     true,
			expectedErr: ErrTest,
		},
		{
			name:   "valid params, error processNotExistTitles(Store)",
			titles: []string{"Bitcoin", "ETC", "TON"},
			setupMock: func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider, mockLogger *mocks.MockLogger) {
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return([]string{"Bitcoin", "ETC", "TON"}, nil)
				mockCryptoProvider.EXPECT().
					GetActualRates(gomock.Any(), []string{"Bitcoin", "ETC", "TON"}).
					Return([]*entities.Coin{
						{Title: "Bitcoin", Cost: 1000},
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}, nil)
				mockStorage.EXPECT().
					Store(gomock.Any(), []*entities.Coin{
						{Title: "Bitcoin", Cost: 1000},
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}).
					Return(nil)
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return([]string{"Bitcoin"}, nil)
				mockCryptoProvider.EXPECT().
					GetActualRates(gomock.Any(), []string{"ETC", "TON"}).
					Return([]*entities.Coin{
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}, nil)
				mockStorage.EXPECT().
					Store(gomock.Any(), []*entities.Coin{
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}).
					Return(ErrTest)
			},
			expectedRes: nil,
			wantErr:     true,
			expectedErr: ErrTest,
		},
		{
			name:   "valid params, response GetAggregateCoins with error",
			titles: []string{"Bitcoin", "ETC", "TON"},
			setupMock: func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider, mockLogger *mocks.MockLogger) {
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return([]string{"Bitcoin", "ETC", "TON"}, nil)
				mockCryptoProvider.EXPECT().
					GetActualRates(gomock.Any(), []string{"Bitcoin", "ETC", "TON"}).
					Return([]*entities.Coin{
						{Title: "Bitcoin", Cost: 1000},
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}, nil)
				mockStorage.EXPECT().
					Store(gomock.Any(), []*entities.Coin{
						{Title: "Bitcoin", Cost: 1000},
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}).
					Return(nil)
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return([]string{"Bitcoin", "ETC", "TON"}, nil)
				mockStorage.EXPECT().
					GetAggregateCoins(gomock.Any(), []string{"Bitcoin", "ETC", "TON"}, gomock.Any()).
					Return(nil, ErrTest)
			},
			expectedRes: nil,
			wantErr:     true,
			expectedErr: ErrTest,
		},
		{
			name:   "invalid params, empty titles",
			titles: []string{},
			setupMock: func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider, mockLogger *mocks.MockLogger) {
			},
			expectedRes: nil,
			wantErr:     true,
			expectedErr: entities.ErrInvalidParam,
		},
		{
			name:   "valid params, error ActualizeRates(GetCoinsList)",
			titles: []string{"Bitcoin", "ETC", "TON"},
			setupMock: func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider, mockLogger *mocks.MockLogger) {
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return(nil, ErrTest)
			},
			expectedRes: nil,
			wantErr:     true,
			expectedErr: ErrTest,
		},
		{
			name:   "valid params, error ActualizeRates(GetActualRates)",
			titles: []string{"Bitcoin", "ETC", "TON"},
			setupMock: func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider, mockLogger *mocks.MockLogger) {
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return([]string{"Bitcoin", "ETC", "TON"}, nil)
				mockCryptoProvider.EXPECT().
					GetActualRates(gomock.Any(), []string{"Bitcoin", "ETC", "TON"}).
					Return(nil, ErrTest)
			},
			expectedRes: nil,
			wantErr:     true,
			expectedErr: ErrTest,
		},
		{
			name:   "valid params, error ActualizeRates(Store)",
			titles: []string{"Bitcoin", "ETC", "TON"},
			setupMock: func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider, mockLogger *mocks.MockLogger) {
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return([]string{"Bitcoin", "ETC", "TON"}, nil)
				mockCryptoProvider.EXPECT().
					GetActualRates(gomock.Any(), []string{"Bitcoin", "ETC", "TON"}).
					Return([]*entities.Coin{
						{Title: "Bitcoin", Cost: 1000},
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}, nil)
				mockStorage.EXPECT().
					Store(gomock.Any(), []*entities.Coin{
						{Title: "Bitcoin", Cost: 1000},
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}).
					Return(ErrTest)
			},
			expectedRes: nil,
			wantErr:     true,
			expectedErr: ErrTest,
		},
	}

	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {

			tc.setupMock(mockStorage, mockCryptoProvider, mockLogger)

			coins, err := service.GetAggregateRates(context.Background(), tc.titles, cases.AggTypeMax)

			if tc.wantErr {
				require.ErrorIs(t, err, tc.expectedErr)
				require.Nil(t, coins)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expectedRes, coins)
		})

	}
}

func TestGetLastRates(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockCryptoProvider := mocks.NewMockCryptoProvider(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	service, _ := cases.NewService(mockCryptoProvider, mockStorage, mockLogger)

	testTable := []struct {
		name        string
		titles      []string
		setupMock   func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider, mockLogger *mocks.MockLogger)
		expectedRes []*entities.Coin
		wantErr     bool
		expectedErr error
	}{
		{
			name:   "valid params, all coins stored",
			titles: []string{"Bitcoin", "ETC", "TON"},
			setupMock: func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider, mockLogger *mocks.MockLogger) {
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return([]string{"Bitcoin", "ETC", "TON"}, nil)
				mockCryptoProvider.EXPECT().
					GetActualRates(gomock.Any(), []string{"Bitcoin", "ETC", "TON"}).
					Return([]*entities.Coin{
						{Title: "Bitcoin", Cost: 1000},
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}, nil)
				mockStorage.EXPECT().
					Store(gomock.Any(), []*entities.Coin{
						{Title: "Bitcoin", Cost: 1000},
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}).
					Return(nil)
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return([]string{"Bitcoin", "ETC", "TON"}, nil)
				mockStorage.EXPECT().
					GetActualCoin(gomock.Any(), []string{"Bitcoin", "ETC", "TON"}).
					Return([]*entities.Coin{
						{Title: "Bitcoin", Cost: 1000},
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}, nil)
			},
			expectedRes: []*entities.Coin{
				{Title: "Bitcoin", Cost: 1000},
				{Title: "ETH", Cost: 5555},
				{Title: "TON", Cost: 1},
			},
			wantErr: false,
		},
		{
			name:   "valid params, not all coins stored",
			titles: []string{"Bitcoin", "ETC", "TON"},
			setupMock: func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider, mockLogger *mocks.MockLogger) {
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return([]string{"Bitcoin", "ETC", "TON"}, nil)
				mockCryptoProvider.EXPECT().
					GetActualRates(gomock.Any(), []string{"Bitcoin", "ETC", "TON"}).
					Return([]*entities.Coin{
						{Title: "Bitcoin", Cost: 1000},
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}, nil)
				mockStorage.EXPECT().
					Store(gomock.Any(), []*entities.Coin{
						{Title: "Bitcoin", Cost: 1000},
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}).
					Return(nil)
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return([]string{"Bitcoin"}, nil)
				mockCryptoProvider.EXPECT().
					GetActualRates(gomock.Any(), []string{"ETC", "TON"}).
					Return([]*entities.Coin{
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}, nil)
				mockStorage.EXPECT().
					Store(gomock.Any(), []*entities.Coin{
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}).
					Return(nil)
				mockStorage.EXPECT().
					GetActualCoin(gomock.Any(), []string{"Bitcoin", "ETC", "TON"}).
					Return([]*entities.Coin{
						{Title: "Bitcoin", Cost: 1000},
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}, nil)
			},
			expectedRes: []*entities.Coin{
				{Title: "Bitcoin", Cost: 1000},
				{Title: "ETH", Cost: 5555},
				{Title: "TON", Cost: 1},
			},
			wantErr: false,
		},
		{
			name:   "valid params, error processNotExistTitles(GetCoinsList)",
			titles: []string{"Bitcoin", "ETC", "TON"},
			setupMock: func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider, mockLogger *mocks.MockLogger) {
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return([]string{"Bitcoin", "ETC", "TON"}, nil)
				mockCryptoProvider.EXPECT().
					GetActualRates(gomock.Any(), []string{"Bitcoin", "ETC", "TON"}).
					Return([]*entities.Coin{
						{Title: "Bitcoin", Cost: 1000},
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}, nil)
				mockStorage.EXPECT().
					Store(gomock.Any(), []*entities.Coin{
						{Title: "Bitcoin", Cost: 1000},
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}).
					Return(nil)
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return(nil, ErrTest)
			},
			expectedRes: nil,
			wantErr:     true,
			expectedErr: ErrTest,
		},
		{
			name:   "valid params, error processNotExistTitles(GetActualRates)",
			titles: []string{"Bitcoin", "ETC", "TON"},
			setupMock: func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider, mockLogger *mocks.MockLogger) {
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return([]string{"Bitcoin", "ETC", "TON"}, nil)
				mockCryptoProvider.EXPECT().
					GetActualRates(gomock.Any(), []string{"Bitcoin", "ETC", "TON"}).
					Return([]*entities.Coin{
						{Title: "Bitcoin", Cost: 1000},
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}, nil)
				mockStorage.EXPECT().
					Store(gomock.Any(), []*entities.Coin{
						{Title: "Bitcoin", Cost: 1000},
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}).
					Return(nil)
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return([]string{"Bitcoin"}, nil)
				mockCryptoProvider.EXPECT().
					GetActualRates(gomock.Any(), []string{"ETC", "TON"}).
					Return(nil, ErrTest)
			},
			expectedRes: nil,
			wantErr:     true,
			expectedErr: ErrTest,
		},
		{
			name:   "valid params, error processNotExistTitles(Store)",
			titles: []string{"Bitcoin", "ETC", "TON"},
			setupMock: func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider, mockLogger *mocks.MockLogger) {
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return([]string{"Bitcoin", "ETC", "TON"}, nil)
				mockCryptoProvider.EXPECT().
					GetActualRates(gomock.Any(), []string{"Bitcoin", "ETC", "TON"}).
					Return([]*entities.Coin{
						{Title: "Bitcoin", Cost: 1000},
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}, nil)
				mockStorage.EXPECT().
					Store(gomock.Any(), []*entities.Coin{
						{Title: "Bitcoin", Cost: 1000},
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}).
					Return(nil)
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return([]string{"Bitcoin"}, nil)
				mockCryptoProvider.EXPECT().
					GetActualRates(gomock.Any(), []string{"ETC", "TON"}).
					Return([]*entities.Coin{
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}, nil)
				mockStorage.EXPECT().
					Store(gomock.Any(), []*entities.Coin{
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}).
					Return(ErrTest)
			},
			expectedRes: nil,
			wantErr:     true,
			expectedErr: ErrTest,
		},
		{
			name:   "valid params, response GetActualCoin with error",
			titles: []string{"Bitcoin", "ETC", "TON"},
			setupMock: func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider, mockLogger *mocks.MockLogger) {
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return([]string{"Bitcoin", "ETC", "TON"}, nil)
				mockCryptoProvider.EXPECT().
					GetActualRates(gomock.Any(), []string{"Bitcoin", "ETC", "TON"}).
					Return([]*entities.Coin{
						{Title: "Bitcoin", Cost: 1000},
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}, nil)
				mockStorage.EXPECT().
					Store(gomock.Any(), []*entities.Coin{
						{Title: "Bitcoin", Cost: 1000},
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}).
					Return(nil)
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return([]string{"Bitcoin", "ETC", "TON"}, nil)
				mockStorage.EXPECT().
					GetActualCoin(gomock.Any(), []string{"Bitcoin", "ETC", "TON"}).
					Return(nil, ErrTest)
			},
			expectedRes: nil,
			wantErr:     true,
			expectedErr: ErrTest,
		},
		{
			name:   "invalid params, empty titles",
			titles: []string{},
			setupMock: func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider, mockLogger *mocks.MockLogger) {
			},
			expectedRes: nil,
			wantErr:     true,
			expectedErr: entities.ErrInvalidParam,
		},
		{
			name:   "valid params, error ActualizeRates(GetCoinsList)",
			titles: []string{"Bitcoin", "ETC", "TON"},
			setupMock: func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider, mockLogger *mocks.MockLogger) {
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return(nil, ErrTest)
			},
			expectedRes: nil,
			wantErr:     true,
			expectedErr: ErrTest,
		},
		{
			name:   "valid params, error ActualizeRates(GetActualRates)",
			titles: []string{"Bitcoin", "ETC", "TON"},
			setupMock: func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider, mockLogger *mocks.MockLogger) {
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return([]string{"Bitcoin", "ETC", "TON"}, nil)
				mockCryptoProvider.EXPECT().
					GetActualRates(gomock.Any(), []string{"Bitcoin", "ETC", "TON"}).
					Return(nil, ErrTest)
			},
			expectedRes: nil,
			wantErr:     true,
			expectedErr: ErrTest,
		},
		{
			name:   "valid params, error ActualizeRates(Store)",
			titles: []string{"Bitcoin", "ETC", "TON"},
			setupMock: func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider, mockLogger *mocks.MockLogger) {
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return([]string{"Bitcoin", "ETC", "TON"}, nil)
				mockCryptoProvider.EXPECT().
					GetActualRates(gomock.Any(), []string{"Bitcoin", "ETC", "TON"}).
					Return([]*entities.Coin{
						{Title: "Bitcoin", Cost: 1000},
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}, nil)
				mockStorage.EXPECT().
					Store(gomock.Any(), []*entities.Coin{
						{Title: "Bitcoin", Cost: 1000},
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}).
					Return(ErrTest)
			},
			expectedRes: nil,
			wantErr:     true,
			expectedErr: ErrTest,
		},
	}
	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {

			tc.setupMock(mockStorage, mockCryptoProvider, mockLogger)

			coins, err := service.GetLastRates(context.Background(), tc.titles)

			if tc.wantErr {
				require.ErrorIs(t, err, tc.expectedErr)
				require.Nil(t, coins)
				return
			}

			require.NoError(t, err)
			require.Equal(t, coins, tc.expectedRes)
		})
	}
}

func TestActualizeRates(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockCryptoProvider := mocks.NewMockCryptoProvider(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	service, _ := cases.NewService(mockCryptoProvider, mockStorage, mockLogger)

	testTable := []struct {
		name        string
		setupMock   func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider, mockLogger *mocks.MockLogger)
		wantErr     bool
		expectedErr error
	}{
		{
			name: "valid params",
			setupMock: func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider, mockLogger *mocks.MockLogger) {
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return([]string{"Bitcoin", "TON", "ETH"}, nil)
				mockCryptoProvider.EXPECT().
					GetActualRates(gomock.Any(), []string{"Bitcoin", "TON", "ETH"}).
					Return([]*entities.Coin{
						{Title: "Bitcoin"},
						{Title: "TON"},
						{Title: "ETH"},
					}, nil)
				mockStorage.EXPECT().
					Store(gomock.Any(), []*entities.Coin{
						{Title: "Bitcoin"},
						{Title: "TON"},
						{Title: "ETH"},
					}).
					Return(nil)
			},
			wantErr:     false,
			expectedErr: nil,
		},
		{
			name: "response GetCoinsList with error",
			setupMock: func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider, mockLogger *mocks.MockLogger) {
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return(nil, ErrTest)
			},
			wantErr:     true,
			expectedErr: ErrTest,
		},
		{
			name: "response GetActualRates with error",
			setupMock: func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider, mockLogger *mocks.MockLogger) {
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return([]string{"Bitcoin", "TON", "ETH"}, nil)
				mockCryptoProvider.EXPECT().
					GetActualRates(gomock.Any(), []string{"Bitcoin", "TON", "ETH"}).
					Return(nil, ErrTest)
			},
			wantErr:     true,
			expectedErr: ErrTest,
		},
		{
			name: "response Store with error",
			setupMock: func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider, mockLogger *mocks.MockLogger) {
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return([]string{"Bitcoin", "TON", "ETH"}, nil)
				mockCryptoProvider.EXPECT().
					GetActualRates(gomock.Any(), []string{"Bitcoin", "TON", "ETH"}).
					Return([]*entities.Coin{
						{Title: "Bitcoin"},
						{Title: "TON"},
						{Title: "ETH"},
					}, nil)
				mockStorage.EXPECT().
					Store(gomock.Any(), []*entities.Coin{
						{Title: "Bitcoin"},
						{Title: "TON"},
						{Title: "ETH"},
					}).
					Return(ErrTest)
			},
			wantErr:     true,
			expectedErr: ErrTest,
		},
	}
	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.setupMock(mockStorage, mockCryptoProvider, mockLogger)

			err := service.ActualizeRates(context.Background())

			if tc.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.expectedErr)
			}
		})
	}
}

func TestNewService(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockCryptoProvider := mocks.NewMockCryptoProvider(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	var nilStorage cases.Storage = nil
	var nilCryptoProvider cases.CryptoProvider = nil
	var nilLogger cases.Logger = nil

	testTable := []struct {
		name           string
		storage        cases.Storage
		cryptoProvider cases.CryptoProvider
		logger         cases.Logger
		wantErr        bool
		expectedErr    error
	}{
		{
			name:           "invalid storage and crypto provider",
			storage:        nil,
			cryptoProvider: nil,
			logger:         mockLogger,
			wantErr:        true,
			expectedErr:    entities.ErrInvalidParam,
		},
		{
			name:           "invalid storage and crypto provider and logger",
			storage:        nil,
			cryptoProvider: nil,
			logger:         nil,
			wantErr:        true,
			expectedErr:    entities.ErrInvalidParam,
		},
		{
			name:           "invalid storage",
			storage:        nil,
			cryptoProvider: mockCryptoProvider,
			logger:         mockLogger,
			wantErr:        true,
			expectedErr:    entities.ErrInvalidParam,
		},
		{
			name:           "invalid logger",
			storage:        mockStorage,
			cryptoProvider: mockCryptoProvider,
			logger:         nil,
			wantErr:        true,
			expectedErr:    entities.ErrInvalidParam,
		},
		{
			name:           "invalid crypto provider",
			storage:        mockStorage,
			cryptoProvider: nil,
			logger:         mockLogger,
			wantErr:        true,
			expectedErr:    entities.ErrInvalidParam,
		},
		{
			name:           "nil interface storage and crypto provider",
			storage:        nilStorage,
			cryptoProvider: nilCryptoProvider,
			logger:         mockLogger,
			wantErr:        true,
			expectedErr:    entities.ErrInvalidParam,
		},
		{
			name:           "nil interface storage and crypto provider and logger",
			storage:        nilStorage,
			cryptoProvider: nilCryptoProvider,
			logger:         nilLogger,
			wantErr:        true,
			expectedErr:    entities.ErrInvalidParam,
		},
		{
			name:           "nil interface storage",
			storage:        nilStorage,
			cryptoProvider: mockCryptoProvider,
			logger:         mockLogger,
			wantErr:        true,
			expectedErr:    entities.ErrInvalidParam,
		},
		{
			name:           "nil interface logger",
			storage:        mockStorage,
			cryptoProvider: mockCryptoProvider,
			logger:         nilLogger,
			wantErr:        true,
			expectedErr:    entities.ErrInvalidParam,
		},
		{
			name:           "nil interface crypto provider",
			storage:        mockStorage,
			cryptoProvider: nilCryptoProvider,
			logger:         mockLogger,
			wantErr:        true,
			expectedErr:    entities.ErrInvalidParam,
		},
		{
			name:           "Valid data",
			storage:        mockStorage,
			cryptoProvider: mockCryptoProvider,
			logger:         mockLogger,
			wantErr:        false,
			expectedErr:    nil,
		},
	}

	//TODO: фикс тестов
	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {

			service, err := cases.NewService(tc.cryptoProvider, tc.storage, tc.logger)

			if tc.wantErr {
				require.ErrorIs(t, err, tc.expectedErr)
				require.Nil(t, service)
				return
			}

			require.NoError(t, err)
		})
	}
}
