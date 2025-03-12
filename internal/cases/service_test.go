package cases_test

import (
	"context"
	"crypto-project/internal/entities"
	"github.com/stretchr/testify/require"
	"testing"

	"go.uber.org/mock/gomock"

	"crypto-project/internal/cases"
	"crypto-project/internal/cases/mocks"
)

func TestAggregateFunctions(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockCryptoProvider := mocks.NewMockCryptoProvider(ctrl)

	service := &cases.Service{
		Storage:  mockStorage,
		Provider: mockCryptoProvider,
	}

	testTable := []struct {
		name        string
		titles      []string
		setupMock   func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider)
		expectedRes []*entities.Coin
		wantErr     bool
		expectedErr error
	}{
		{
			name:   "valid params, all coins stored",
			titles: []string{"Bitcoin", "ETC", "TON"},
			setupMock: func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider) {
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return([]string{"Bitcoin", "ETC", "TON"}, nil).
					Times(3)
				mockStorage.EXPECT().
					GetAggregateCoins(gomock.Any(), []string{"Bitcoin", "ETC", "TON"}, gomock.Any()).
					Return([]*entities.Coin{
						{Title: "Bitcoin", Cost: 1000},
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}, nil).
					Times(3)
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
			setupMock: func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider) {
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return([]string{"Bitcoin"}, nil).
					Times(3)
				mockCryptoProvider.EXPECT().
					GetActualRates(gomock.Any(), []string{"ETC", "TON"}).
					Return([]*entities.Coin{
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}, nil).
					Times(3)
				mockStorage.EXPECT().
					Store(gomock.Any(), []*entities.Coin{
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}).
					Return(nil).
					Times(3)
				mockStorage.EXPECT().
					GetAggregateCoins(gomock.Any(), []string{"Bitcoin", "ETC", "TON"}, gomock.Any()).
					Return([]*entities.Coin{
						{Title: "Bitcoin", Cost: 1000},
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}, nil).
					Times(3)
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
			setupMock: func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider) {
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return(nil, entities.ErrStorage).
					Times(3)
			},
			expectedRes: nil,
			wantErr:     true,
			expectedErr: entities.ErrStorage,
		},
		{
			name:   "valid params, error processNotExistTitles(GetActualRates)",
			titles: []string{"Bitcoin", "ETC", "TON"},
			setupMock: func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider) {
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return([]string{"Bitcoin"}, nil).
					Times(3)
				mockCryptoProvider.EXPECT().
					GetActualRates(gomock.Any(), []string{"ETC", "TON"}).
					Return(nil, entities.ErrProvider).
					Times(3)
			},
			expectedRes: nil,
			wantErr:     true,
			expectedErr: entities.ErrProvider,
		},
		{
			name:   "valid params, error processNotExistTitles(Store)",
			titles: []string{"Bitcoin", "ETC", "TON"},
			setupMock: func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider) {
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return([]string{"Bitcoin"}, nil).
					Times(3)
				mockCryptoProvider.EXPECT().
					GetActualRates(gomock.Any(), []string{"ETC", "TON"}).
					Return([]*entities.Coin{
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}, nil).
					Times(3)
				mockStorage.EXPECT().
					Store(gomock.Any(), []*entities.Coin{
						{Title: "ETH", Cost: 5555},
						{Title: "TON", Cost: 1},
					}).
					Return(entities.ErrStorage).
					Times(3)
			},
			expectedRes: nil,
			wantErr:     true,
			expectedErr: entities.ErrStorage,
		},
		{
			name:   "valid params, response GetAggregateCoins with error",
			titles: []string{"Bitcoin", "ETC", "TON"},
			setupMock: func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider) {
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return([]string{"Bitcoin", "ETC", "TON"}, nil).
					Times(3)
				mockStorage.EXPECT().
					GetAggregateCoins(gomock.Any(), []string{"Bitcoin", "ETC", "TON"}, gomock.Any()).
					Return(nil, entities.ErrStorage).
					Times(3)
			},
			expectedRes: nil,
			wantErr:     true,
			expectedErr: entities.ErrStorage,
		},
		{
			name:        "empty titles",
			titles:      []string{},
			setupMock:   func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider) {},
			expectedRes: nil,
			wantErr:     true,
			expectedErr: entities.ErrInvalidParam,
		},
	}

	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {

			tc.setupMock(mockStorage, mockCryptoProvider)

			maxCoins, errMax := service.GetMaxRates(context.Background(), tc.titles)
			minCoins, errMin := service.GetMinRates(context.Background(), tc.titles)
			avgCoins, errAvg := service.GetAvgRates(context.Background(), tc.titles)

			if tc.wantErr {
				require.ErrorIs(t, errMax, tc.expectedErr)
				require.ErrorIs(t, errMin, tc.expectedErr)
				require.ErrorIs(t, errAvg, tc.expectedErr)
				require.Nil(t, maxCoins)
				require.Nil(t, minCoins)
				require.Nil(t, avgCoins)
				return
			}

			require.NoError(t, errMax)
			require.NoError(t, errMin)
			require.NoError(t, errAvg)
			require.Equal(t, tc.expectedRes, maxCoins)
			require.Equal(t, tc.expectedRes, minCoins)
			require.Equal(t, tc.expectedRes, avgCoins)
		})

	}
}

func TestGetLastRates(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockCryptoProvider := mocks.NewMockCryptoProvider(ctrl)

	service := &cases.Service{
		Storage:  mockStorage,
		Provider: mockCryptoProvider,
	}

	testTable := []struct {
		name        string
		titles      []string
		setupMock   func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider)
		expectedRes []*entities.Coin
		wantErr     bool
		expectedErr error
	}{
		{
			name:   "valid params, all coins stored",
			titles: []string{"Bitcoin", "ETC", "TON"},
			setupMock: func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider) {
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
			setupMock: func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider) {
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
			setupMock: func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider) {
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return(nil, entities.ErrStorage)
			},
			expectedRes: nil,
			wantErr:     true,
			expectedErr: entities.ErrStorage,
		},
		{
			name:   "valid params, error processNotExistTitles(GetActualRates)",
			titles: []string{"Bitcoin", "ETC", "TON"},
			setupMock: func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider) {
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return([]string{"Bitcoin"}, nil)
				mockCryptoProvider.EXPECT().
					GetActualRates(gomock.Any(), []string{"ETC", "TON"}).
					Return(nil, entities.ErrProvider)
			},
			expectedRes: nil,
			wantErr:     true,
			expectedErr: entities.ErrProvider,
		},
		{
			name:   "valid params, error processNotExistTitles(Store)",
			titles: []string{"Bitcoin", "ETC", "TON"},
			setupMock: func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider) {
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
					Return(entities.ErrStorage)
			},
			expectedRes: nil,
			wantErr:     true,
			expectedErr: entities.ErrStorage,
		},
		{
			name:   "valid params, response GetActualCoin with error",
			titles: []string{"Bitcoin", "ETC", "TON"},
			setupMock: func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider) {
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return([]string{"Bitcoin", "ETC", "TON"}, nil)
				mockStorage.EXPECT().
					GetActualCoin(gomock.Any(), []string{"Bitcoin", "ETC", "TON"}).
					Return(nil, entities.ErrStorage)
			},
			expectedRes: nil,
			wantErr:     true,
			expectedErr: entities.ErrStorage,
		},
		{
			name:        "empty titles",
			titles:      []string{},
			setupMock:   func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider) {},
			expectedRes: nil,
			wantErr:     true,
			expectedErr: entities.ErrInvalidParam,
		},
	}
	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {

			tc.setupMock(mockStorage, mockCryptoProvider)

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

	service := &cases.Service{
		Storage:  mockStorage,
		Provider: mockCryptoProvider,
	}

	testTable := []struct {
		name        string
		setupMock   func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider)
		wantErr     bool
		expectedErr error
	}{
		{
			name: "valid params",
			setupMock: func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider) {
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
			setupMock: func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider) {
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return(nil, entities.ErrStorage)
			},
			wantErr:     true,
			expectedErr: entities.ErrStorage,
		},
		{
			name: "response GetActualRates with error",
			setupMock: func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider) {
				mockStorage.EXPECT().
					GetCoinsList(gomock.Any()).
					Return([]string{"Bitcoin", "TON", "ETH"}, nil)
				mockCryptoProvider.EXPECT().
					GetActualRates(gomock.Any(), []string{"Bitcoin", "TON", "ETH"}).
					Return(nil, entities.ErrStorage)
			},
			wantErr:     true,
			expectedErr: entities.ErrStorage,
		},
		{
			name: "response Store with error",
			setupMock: func(mockStorage *mocks.MockStorage, mockCryptoProvider *mocks.MockCryptoProvider) {
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
					Return(entities.ErrStorage)
			},
			wantErr:     true,
			expectedErr: entities.ErrStorage,
		},
	}
	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.setupMock(mockStorage, mockCryptoProvider)

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

	var nilStorage cases.Storage = nil
	var nilCryptoProvider cases.CryptoProvider = nil

	testTable := []struct {
		name           string
		storage        cases.Storage
		cryptoProvider cases.CryptoProvider
		wantErr        bool
		expectedErr    error
	}{
		{
			name:           "invalid storage and crypto provider",
			storage:        nil,
			cryptoProvider: nil,
			wantErr:        true,
			expectedErr:    entities.ErrInvalidParam,
		},
		{
			name:           "invalid storage",
			storage:        nil,
			cryptoProvider: mockCryptoProvider,
			wantErr:        true,
			expectedErr:    entities.ErrInvalidParam,
		},
		{
			name:           "invalid crypto provider",
			storage:        mockStorage,
			cryptoProvider: nil,
			wantErr:        true,
			expectedErr:    entities.ErrInvalidParam,
		},
		{
			name:           "nil interface storage and crypto provider",
			storage:        nilStorage,
			cryptoProvider: nilCryptoProvider,
			wantErr:        true,
			expectedErr:    entities.ErrInvalidParam,
		},
		{
			name:           "nil interface storage",
			storage:        nilStorage,
			cryptoProvider: nil,
			wantErr:        true,
			expectedErr:    entities.ErrInvalidParam,
		},
		{
			name:           "nil interface crypto provider",
			storage:        nil,
			cryptoProvider: nilCryptoProvider,
			wantErr:        true,
			expectedErr:    entities.ErrInvalidParam,
		},
		{
			name:           "Valid data",
			storage:        mockStorage,
			cryptoProvider: mockCryptoProvider,
			wantErr:        false,
			expectedErr:    nil,
		},
	}

	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {

			service, err := cases.NewService(tc.cryptoProvider, tc.storage)

			if tc.wantErr {
				require.ErrorIs(t, err, tc.expectedErr)
				require.Nil(t, service)
				return
			}

			require.NoError(t, err)
			require.Equal(t, &cases.Service{
				Storage:  tc.storage,
				Provider: tc.cryptoProvider,
			}, service)
		})
	}
}
