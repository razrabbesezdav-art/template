package finish_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/razrabbesezdav-art/template/internal/domain/trip"
	"github.com/razrabbesezdav-art/template/internal/usecase/trip/finish"
)

func TestUseCase_Finish(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMocktripRepository(ctrl)
	mockTx := NewMocktxManager(ctrl)
	uc := finish.New(mockRepo, mockTx)

	id := uuid.New()
	
	t.Run("success", func(t *testing.T) {
		activeTrip := trip.Restore(id, uuid.New(), uuid.New(), 0, 0, 0, 0, 100, trip.StatusActive, time.Now(), nil)

		mockTx.EXPECT().Do(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(ctx context.Context) error) error {
				return fn(ctx)
			},
		)
		mockRepo.EXPECT().Get(gomock.Any(), id).Return(activeTrip, nil)
		mockRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

		res, err := uc.Finish(context.Background(), id)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Status() != trip.StatusCompleted {
			t.Errorf("expected status completed, got %v", res.Status())
		}
	})

	t.Run("already completed", func(t *testing.T) {
		completedTrip := trip.Restore(id, uuid.New(), uuid.New(), 0, 0, 0, 0, 100, trip.StatusCompleted, time.Now(), nil)

		mockTx.EXPECT().Do(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(ctx context.Context) error) error {
				return fn(ctx)
			},
		)
		mockRepo.EXPECT().Get(gomock.Any(), id).Return(completedTrip, nil)
		
		_, err := uc.Finish(context.Background(), id)
		if !errors.Is(err, trip.ErrTripCompleted) {
			t.Fatalf("expected ErrTripCompleted, got %v", err)
		}
	})
}