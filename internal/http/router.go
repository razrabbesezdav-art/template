package http

import (
	"net/http"

	api "github.com/razrabbesezdav-art/template/internal/generated"
	"github.com/razrabbesezdav-art/template/internal/http/trip/create"
	"github.com/razrabbesezdav-art/template/internal/http/trip/finish"
	"github.com/razrabbesezdav-art/template/internal/http/trip/get"
)

type Router struct {
	create *create.Handler
	get    *get.Handler
	finish *finish.Handler
}

func NewRouter(
	createHandler *create.Handler,
	getHandler *get.Handler,
	finishHandler *finish.Handler,
) *Router {
	return &Router{
		create: createHandler,
		get:    getHandler,
		finish: finishHandler,
	}
}

func (r *Router) CreateTrip(w http.ResponseWriter, req *http.Request, _ api.CreateTripParams) {
	r.create.CreateTrip(w, req)
}

func (r *Router) GetTrip(w http.ResponseWriter, req *http.Request, tripId api.TripId) {
	r.get.GetTrip(w, req)
}

func (r *Router) FinishTrip(w http.ResponseWriter, req *http.Request, tripId api.TripId) {
	r.finish.FinishTrip(w, req)
}

func (r *Router) ListTripPositions(w http.ResponseWriter, req *http.Request, tripId api.TripId) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (r *Router) CreateTripPosition(w http.ResponseWriter, req *http.Request, tripId api.TripId) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (r *Router) Health(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (r *Router) Ready(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}