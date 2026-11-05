package create_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/razrabbesezdav-art/template/internal/domain/trip"
	"github.com/razrabbesezdav-art/template/internal/usecase/trip/create"
)

func TestUseCase_Create(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(repo *MocktripRepository, tx *MocktxManager)
		wantErr   error
	}{
		{
			name: "success",
			mockSetup: func(repo *MocktripRepository, tx *MocktxManager) {
				tx.EXPECT().Do(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(ctx context.Context) error) error {
						return fn(ctx)
					},
				)
				repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "driver busy error",
			mockSetup: func(repo *MocktripRepository, tx *MocktxManager) {
				tx.EXPECT().Do(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(ctx context.Context) error) error {
						return fn(ctx)
					},
				)
				repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(trip.ErrDriverBusy)
			},
			wantErr: trip.ErrDriverBusy,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := NewMocktripRepository(ctrl)
			mockTx := NewMocktxManager(ctrl)

			tt.mockSetup(mockRepo, mockTx)

			// Используем create.New, так как сам UseCase лежит в пакете create
			uc := create.New(mockRepo, mockTx)

			_, err := uc.Create(context.Background(), create.Request{
				UserID:   uuid.New(),
				DriverID: uuid.New(),
				StartLat: 0, StartLon: 0,
				EndLat: 0, EndLon: 0,
				Price: 100,
			})

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}