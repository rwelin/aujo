package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
)

func Err(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(err.Error()))
	fmt.Fprintln(os.Stderr, err)
}

type Callbacks interface {
	Mix() ([]byte, error)
	UpdateInstrumentHarmonics(instrument int, harmonics []float64) error
}

type handler struct {
	Callbacks Callbacks
}

func (h *handler) handleInstrumentPut(w http.ResponseWriter, r *http.Request) {
	var voice []float64
	err := json.NewDecoder(r.Body).Decode(&voice)
	if err != nil {
		Err(w, err)
		return
	}

	id := mux.Vars(r)["id"]
	i, err := strconv.Atoi(id)
	if err != nil {
		Err(w, err)
	}

	if err := h.Callbacks.UpdateInstrumentHarmonics(i, voice); err != nil {
		Err(w, err)
	}
}

func (h *handler) handleMixGet(w http.ResponseWriter, r *http.Request) {
	d, err := h.Callbacks.Mix()
	if err != nil {
		Err(w, err)
		return
	}

	w.Write(d)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if r.Method == http.MethodOptions {
			return
		}
		next.ServeHTTP(w, r)
	})
}

func NewHandler(cb Callbacks) http.Handler {
	h := &handler{
		Callbacks: cb,
	}

	sr := mux.NewRouter()
	sr.HandleFunc("/instruments/{id}", h.handleInstrumentPut).Methods(http.MethodPut)
	sr.HandleFunc("/mix", h.handleMixGet).Methods(http.MethodGet)

	r := mux.NewRouter()
	r.Use(corsMiddleware)
	r.PathPrefix("/").Handler(sr)
	return r
}
