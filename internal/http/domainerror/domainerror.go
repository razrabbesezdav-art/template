package domainerror

import (
	"errors"
	"net/http"

	"github.com/razrabbesezdav-art/template/internal/domain/trip"
	api "github.com/razrabbesezdav-art/template/internal/generated"
)

func Problem(err error) (int, api.Problem) {
	switch {
	case errors.Is(err, trip.ErrTripNotFound):
		return http.StatusNotFound, api.Problem{
			Code:   "trip_not_found",
			Title:  "Trip not found",
			Status: http.StatusNotFound,
			Type:   "about:blank",
		}
	case errors.Is(err, trip.ErrTripCompleted):
		return http.StatusConflict, api.Problem{
			Code:   "trip_completed",
			Title:  "Trip is already completed",
			Status: http.StatusConflict,
			Type:   "about:blank",
		}
	case errors.Is(err, trip.ErrDriverBusy):
		return http.StatusConflict, api.Problem{
			Code:   "driver_busy",
			Title:  "Driver already has an active trip",
			Status: http.StatusConflict,
			Type:   "about:blank",
		}
	default:
		return http.StatusInternalServerError, api.Problem{
			Code:   "internal_error",
			Title:  "Internal server error",
			Status: http.StatusInternalServerError,
			Type:   "about:blank",
		}
	}
}