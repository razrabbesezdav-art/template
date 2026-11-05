package get

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	api "github.com/razrabbesezdav-art/template/internal/generated"
	"github.com/razrabbesezdav-art/template/internal/http/domainerror"
	usecase "github.com/razrabbesezdav-art/template/internal/usecase/trip/get"
)

type Handler struct {
	uc *usecase.UseCase
}

func NewHandler(uc *usecase.UseCase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) GetTrip(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "tripId")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid trip id", http.StatusBadRequest)
		return
	}

	t, err := h.uc.Get(r.Context(), id)
	if err != nil {
		status, prob := domainerror.Problem(err)
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(prob)
		return
	}

	resp := api.Trip{
		Id:       t.ID(),
		UserId:   t.UserID(),
		DriverId: t.DriverID(),
		StartPoint: api.Coordinates{Latitude: t.StartLat(), Longitude: t.StartLon()},
		EndPoint:   api.Coordinates{Latitude: t.EndLat(), Longitude: t.EndLon()},
		Price:      t.Price(),
		Status:     api.TripStatus(t.Status()),
		StartedAt:  t.StartedAt(),
	}
	if t.FinishedAt() != nil {
		resp.FinishedAt = t.FinishedAt()
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}