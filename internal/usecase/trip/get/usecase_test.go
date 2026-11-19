package get_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/razrabbesezdav-art/template/internal/domain/trip"
	"github.com/razrabbesezdav-art/template/internal/usecase/trip/get"
)

func TestUseCase_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMocktripRepository(ctrl)
	uc := get.New(mockRepo)

	id := uuid.New()

	t.Run("success", func(t *testing.T) {
		expectedTrip := trip.Restore(id, uuid.New(), uuid.New(), 0, 0, 0, 0, 100, trip.StatusActive, time.Now(), nil)
		mockRepo.EXPECT().Get(gomock.Any(), id).Return(expectedTrip, nil)

		_, err := uc.Get(context.Background(), id)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo.EXPECT().Get(gomock.Any(), id).Return(nil, trip.ErrTripNotFound)

		_, err := uc.Get(context.Background(), id)
		if !errors.Is(err, trip.ErrTripNotFound) {
			t.Fatalf("expected ErrTripNotFound, got %v", err)
		}
	})
}