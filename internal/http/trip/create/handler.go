package create

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/google/uuid"

	api "github.com/razrabbesezdav-art/template/internal/generated"
	"github.com/razrabbesezdav-art/template/internal/http/domainerror"
	usecase "github.com/razrabbesezdav-art/template/internal/usecase/trip/create"
)

type Handler struct {
	uc *usecase.UseCase
}

func NewHandler(uc *usecase.UseCase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) CreateTrip(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req api.TripData
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		writeProblem(w, http.StatusBadRequest, "invalid_request", "Request must contain one JSON object")
		return
	}

	if req.Price <= 0 || req.UserId == uuid.Nil || req.DriverId == uuid.Nil {
		writeProblem(w, http.StatusBadRequest, "invalid_request", "Invalid trip data")
		return
	}

	t, err := h.uc.Create(r.Context(), usecase.Request{
		UserID:   req.UserId,
		DriverID: req.DriverId,
		StartLat: req.StartPoint.Latitude,
		StartLon: req.StartPoint.Longitude,
		EndLat:   req.EndPoint.Latitude,
		EndLon:   req.EndPoint.Longitude,
		Price:    req.Price,
	})
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

	w.Header().Set("Location", "/api/v1/trips/"+t.ID().String())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func writeProblem(w http.ResponseWriter, status int, code, title string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(api.Problem{
		Type:   "about:blank",
		Title:  title,
		Status: int32(status),
		Code:   code,
	})
}