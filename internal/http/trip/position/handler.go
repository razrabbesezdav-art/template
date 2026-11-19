package position

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	api "github.com/razrabbesezdav-art/template/internal/generated"
	"github.com/razrabbesezdav-art/template/internal/http/domainerror"
	usecaseSave "github.com/razrabbesezdav-art/template/internal/usecase/trip/position"
)

type Handler struct {
	saveUC *usecaseSave.UseCase
	getUC  *usecaseSave.GetUseCase
}

func NewHandler(saveUC *usecaseSave.UseCase, getUC *usecaseSave.GetUseCase) *Handler {
	return &Handler{saveUC: saveUC, getUC: getUC}
}

func (h *Handler) CreateTripPosition(w http.ResponseWriter, r *http.Request) {
	tripID, ok := parseTripID(w, r)
	if !ok {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req api.PositionData
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

	p, err := h.saveUC.Save(r.Context(), usecaseSave.Request{
		TripID:     tripID,
		Latitude:   req.Coordinates.Latitude,
		Longitude:  req.Coordinates.Longitude,
		RecordedAt: req.RecordedAt,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}

	resp := api.TripPosition{
		Id:     p.ID(),
		TripId: p.TripID(),
		Coordinates: api.Coordinates{
			Latitude:  p.Latitude(),
			Longitude: p.Longitude(),
		},
		RecordedAt: p.RecordedAt(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) ListTripPositions(w http.ResponseWriter, r *http.Request) {
	tripID, ok := parseTripID(w, r)
	if !ok {
		return
	}

	positions, err := h.getUC.GetPositions(r.Context(), tripID)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	resp := api.TripPositionsResponse{Positions: make([]api.TripPosition, 0, len(positions))}
	for _, p := range positions {
		resp.Positions = append(resp.Positions, api.TripPosition{
			Id:     p.ID(),
			TripId: p.TripID(),
			Coordinates: api.Coordinates{
				Latitude:  p.Latitude(),
				Longitude: p.Longitude(),
			},
			RecordedAt: p.RecordedAt(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func parseTripID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	idStr := chi.URLParam(r, "tripId")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "invalid_request", "Invalid trip id")
		return uuid.Nil, false
	}
	return id, true
}

func writeDomainError(w http.ResponseWriter, err error) {
	status, prob := domainerror.Problem(err)
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(prob)
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

var _ = time.Time{}