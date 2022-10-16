package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/a-romancev/crud_task/auth"
	"github.com/a-romancev/crud_task/company"
	"github.com/a-romancev/crud_task/internal/event"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const bodySizeLimit = 1000

type Handler struct {
	router        http.Handler
	crud          *company.CRUD
	eventProducer event.Producer
	pk            *auth.PublicKey
}

func NewHandler(crud *company.CRUD, producer event.Producer, pk *auth.PublicKey) *Handler {
	h := Handler{
		crud:          crud,
		eventProducer: producer,
		pk:            pk,
	}
	r := http.NewServeMux()

	r.HandleFunc("/health", health)
	r.HandleFunc("/v1/companies", h.companies)
	r.HandleFunc("/v1/companies/{ID}", h.companiesByID)

	h.router = r
	return &h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	defer func() {
		if err := recover(); err != nil {
			log.Ctx(ctx).Error().Interface("error", err).Msg("Recovered server error.")
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()
	h.router.ServeHTTP(w, r.WithContext(ctx))
}

func health(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("ok"))
}

func (h *Handler) companies(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:

		var claims auth.APIClaims
		err := h.pk.Verify(auth.Token(r), &claims)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		var cmp company.Company
		err = json.NewDecoder(io.LimitReader(r.Body, bodySizeLimit)).Decode(&cmp)
		if err != nil {
			log.Ctx(r.Context()).Error().Err(err).Msg("Invalid company data.")
			newAPIError(http.StatusBadRequest, "Invalid company data.", err).Write(w)
			return
		}
		cmp.ID = uuid.New()
		err = cmp.Validate()
		if err != nil {
			log.Ctx(r.Context()).Error().Err(err).Msg("Invalid company data.")
			newAPIError(http.StatusBadRequest, "Invalid company data", err).Write(w)
			return
		}
		created, err := h.crud.Repo.Create(r.Context(), cmp)
		if err != nil {
			log.Ctx(r.Context()).Error().Err(err).Msg("Company creation failed.")
			newAPIError(http.StatusBadRequest, "Company creation failed.", err).Write(w)
			return
		}
		err = h.report(cmp)
		if err != nil {
			log.Ctx(r.Context()).Error().Err(err).Msg("Failed to report event.")
		}
		apiResponse{Code: http.StatusCreated, Body: created}.Write(w)
	case http.MethodGet:
		fetched, err := h.crud.Repo.Fetch(context.Background(), company.Lookup{})
		if err != nil {
			log.Ctx(r.Context()).Error().Err(err).Msg("Failed to fetch companies.")
			newAPIError(http.StatusBadRequest, "Failed to fetch companies.", err).Write(w)
			return
		}
		apiResponse{Code: http.StatusOK, Body: fetched}.Write(w)
	default:
		log.Ctx(r.Context()).Error().Str("method", r.Method).Msg("Http method not allowed.")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func (h *Handler) companiesByID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		fetched, err := h.crud.Repo.FetchOne(context.Background(), company.Lookup{})
		if err != nil {
			log.Ctx(r.Context()).Error().Err(err).Msg("Failed to fetch companies.")
			newAPIError(http.StatusBadRequest, "Failed to fetch companies.", err).Write(w)
			return
		}
		apiResponse{Code: http.StatusOK, Body: fetched}.Write(w)
	case http.MethodPut:
		var cmp company.Company
		cmp.ID = uuid.New()
		err := json.NewDecoder(io.LimitReader(r.Body, bodySizeLimit)).Decode(&cmp)
		if err != nil {
			log.Ctx(r.Context()).Error().Err(err).Msg("Invalid company data.")
			newAPIError(http.StatusBadRequest, "Invalid company data", err).Write(w)
			return
		}
		err = cmp.Validate()
		if err != nil {
			log.Ctx(r.Context()).Error().Err(err).Msg("Invalid company data.")
			newAPIError(http.StatusBadRequest, "Invalid company data", err).Write(w)
			return
		}
		created, err := h.crud.Repo.UpdateOne(r.Context(), company.Lookup{}, cmp)
		if err != nil {
			log.Ctx(r.Context()).Error().Err(err).Msg("Company creation failed.")
			newAPIError(http.StatusBadRequest, "Company creation failed.", err).Write(w)
			return
		}
		err = h.report(cmp)
		if err != nil {
			log.Ctx(r.Context()).Error().Err(err).Msg("Failed to report event.")
		}
		apiResponse{Code: http.StatusCreated, Body: created}.Write(w)
	case http.MethodDelete:
		err := h.crud.Repo.DeleteOne(r.Context(), company.Lookup{})
		if err != nil {
			log.Ctx(r.Context()).Error().Err(err).Msg("Company creation failed.")
			newAPIError(http.StatusBadRequest, "Company creation failed.", err).Write(w)
			return
		}
		cmp := company.Company{}
		err = h.report(cmp)
		if err != nil {
			log.Ctx(r.Context()).Error().Err(err).Msg("Failed to report event.")
		}
		apiResponse{Code: http.StatusOK, Body: ""}.Write(w)
	default:
		log.Ctx(r.Context()).Error().Str("method", r.Method).Msg("Http method not allowed.")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *Handler) report(event company.Company) error {
	return nil
	cmpEvent, err := json.Marshal(event)
	if err != nil {
		return err
	}
	err = h.eventProducer.Produce(string(cmpEvent))
	if err != nil {
		return err
	}
	return nil
}

type errBody struct {
	Msg   string `json:"msg"`
	Error string `json:"error"`
}

func newAPIError(code int, msg string, err error) apiResponse {
	return apiResponse{Code: code, Body: errBody{Msg: msg, Error: err.Error()}}
}

// apiResponse is used as a convenient wrapper to send responses.
type apiResponse struct {
	Code int
	Body interface{}
}

func (r apiResponse) Write(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.Code)
	_ = json.NewEncoder(w).Encode(r.Body)
}
