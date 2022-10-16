package main

import (
	"context"
	"encoding/json"
	"github.com/a-romancev/crud_task/auth"
	"github.com/a-romancev/crud_task/company"
	"github.com/a-romancev/crud_task/internal/event"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
)

const bodySizeLimit = 1000

type Repo interface {
	Create(ctx context.Context, request company.Company) (company.Company, error)
	Fetch(ctx context.Context, lookup company.Lookup) ([]company.Company, error)
	FetchOne(ctx context.Context, lookup company.Lookup) (company.Company, error)
	UpdateOne(ctx context.Context, lookup company.Lookup, request company.Company) (company.Company, error)
	DeleteOne(ctx context.Context, lookup company.Lookup) error
}
type Producer interface {
	Produce(m string) error
}

type Handler struct {
	router        http.Handler
	repo          Repo
	eventProducer Producer
	pk            *auth.PublicKey
}

func NewHandler(repo Repo, producer event.Producer, pk *auth.PublicKey) *Handler {
	h := Handler{
		repo:          repo,
		eventProducer: producer,
		pk:            pk,
	}
	r := httprouter.New()
	r.GET("/health", health)
	r.GET("/v1/companies", h.fetchCompanies)
	r.POST("/v1/companies", h.createCompany)
	r.GET("/v1/companies/:id", h.fetchCompany)
	r.PATCH("/v1/companies/:id", h.updateCompany)
	r.DELETE("/v1/companies/:id", h.deleteCompany)

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

func health(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	_, _ = w.Write([]byte("ok"))
}

func (h *Handler) fetchCompanies(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fetched, err := h.repo.Fetch(context.Background(), company.Lookup{})
	if err != nil {
		log.Ctx(r.Context()).Error().Err(err).Msg("Failed to fetch companies.")
		newAPIError(http.StatusBadRequest, "Failed to fetch companies.", err).Write(w)
		return
	}
	apiResponse{Code: http.StatusOK, Body: fetched}.Write(w)
}

func (h *Handler) createCompany(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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
	created, err := h.repo.Create(r.Context(), cmp)
	if err != nil {
		log.Ctx(r.Context()).Error().Err(err).Msg("Company creation failed.")
		newAPIError(http.StatusBadRequest, "Company creation failed.", err).Write(w)
		return
	}
	err = h.report(cmp, "added")
	if err != nil {
		log.Ctx(r.Context()).Error().Err(err).Msg("Failed to report event.")
	}
	apiResponse{Code: http.StatusCreated, Body: created}.Write(w)
}

func (h *Handler) fetchCompany(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")
	uid, err := uuid.Parse(id)
	fetched, err := h.repo.FetchOne(context.Background(), company.Lookup{ID: uid})
	if err != nil {
		log.Ctx(r.Context()).Error().Err(err).Msg("Failed to fetch companies.")
		newAPIError(http.StatusBadRequest, "Failed to fetch companies.", err).Write(w)
		return
	}
	apiResponse{Code: http.StatusOK, Body: fetched}.Write(w)
}

func (h *Handler) deleteCompany(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")
	uid, err := uuid.Parse(id)

	var claims auth.APIClaims
	err = h.pk.Verify(auth.Token(r), &claims)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err = h.repo.DeleteOne(r.Context(), company.Lookup{ID: uid})
	if err != nil {
		log.Ctx(r.Context()).Error().Err(err).Msg("Company creation failed.")
		newAPIError(http.StatusBadRequest, "Company creation failed.", err).Write(w)
		return
	}
	cmp := company.Company{ID: uid}
	err = h.report(cmp, "deleted")
	if err != nil {
		log.Ctx(r.Context()).Error().Err(err).Msg("Failed to report event.")
	}
	apiResponse{Code: http.StatusOK, Body: ""}.Write(w)
}

func (h *Handler) updateCompany(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")
	uid, err := uuid.Parse(id)

	var claims auth.APIClaims
	err = h.pk.Verify(auth.Token(r), &claims)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var cmp company.Company
	err = json.NewDecoder(io.LimitReader(r.Body, bodySizeLimit)).Decode(&cmp)
	if err != nil {
		log.Ctx(r.Context()).Error().Err(err).Msg("Invalid company data.")
		newAPIError(http.StatusBadRequest, "Invalid company data", err).Write(w)
		return
	}
	cmp.ID = uid
	err = cmp.Validate()
	if err != nil {
		log.Ctx(r.Context()).Error().Err(err).Msg("Invalid company data.")
		newAPIError(http.StatusBadRequest, "Invalid company data", err).Write(w)
		return
	}
	created, err := h.repo.UpdateOne(r.Context(), company.Lookup{ID: uid}, cmp)
	if err != nil {
		log.Ctx(r.Context()).Error().Err(err).Msg("Company creation failed.")
		newAPIError(http.StatusBadRequest, "Company creation failed.", err).Write(w)
		return
	}
	err = h.report(cmp, "updated")
	if err != nil {
		log.Ctx(r.Context()).Error().Err(err).Msg("Failed to report event.")
	}
	apiResponse{Code: http.StatusCreated, Body: created}.Write(w)
}

func (h *Handler) report(event company.Company, ev string) error {
	cmpEvent, err := json.Marshal(event)
	if err != nil {
		return err
	}
	err = h.eventProducer.Produce(ev + string(cmpEvent))
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
