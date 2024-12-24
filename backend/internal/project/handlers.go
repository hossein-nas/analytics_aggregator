package project

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hossein-nas/analytics_aggregator/internal/auth"
	"github.com/hossein-nas/analytics_aggregator/internal/project/models"
	"github.com/hossein-nas/analytics_aggregator/pkg/responses"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func RegisterRoutes(router *mux.Router, h *Handler) {
	projectRouter := router.PathPrefix("/projects").Subrouter()

	projectRouter.HandleFunc("", h.CreateProject).Methods(http.MethodPost)
	projectRouter.HandleFunc("", h.ListProjects).Methods(http.MethodGet)
	projectRouter.HandleFunc("/{key}", h.GetProject).Methods(http.MethodGet)
	projectRouter.HandleFunc("/{key}", h.UpdateProject).Methods(http.MethodPut)
	projectRouter.HandleFunc("/{key}/metrics", h.GetMetrics).Methods(http.MethodGet)
}

func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var input models.CreateProjectInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		responses.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	user, ok := auth.GetUserFromContext(r.Context())

	if !ok {
		responses.RespondWithError(w, http.StatusUnauthorized, "There is no user.")
		log.Fatal("There is no user")
		return
	}

	project, err := h.service.CreateProject(r.Context(), user.UserID.Hex(), input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(project)
}

func (h *Handler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	var input models.UpdateProjectInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	project, err := h.service.UpdateProject(r.Context(), key, input)
	if err != nil {
		status := http.StatusInternalServerError
		if err == ErrProjectNotFound {
			status = http.StatusNotFound
		}
		http.Error(w, err.Error(), status)
		return
	}

	json.NewEncoder(w).Encode(project)
}

func (h *Handler) GetProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	project, err := h.service.GetProject(r.Context(), key)
	if err != nil {
		status := http.StatusInternalServerError
		if err == ErrProjectNotFound {
			status = http.StatusNotFound
		}
		http.Error(w, err.Error(), status)
		return
	}

	json.NewEncoder(w).Encode(project)
}

func (h *Handler) ListProjects(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r.Context())

	if !ok {
		log.Fatal("There is no user")
	}

	projects, err := h.service.ListProjects(r.Context(), user.UserID.Hex())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(projects)
}

func (h *Handler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	_, err := h.service.GetProject(r.Context(), key)
	if err != nil {
		status := http.StatusInternalServerError
		if err == ErrProjectNotFound {
			status = http.StatusNotFound
		}
		http.Error(w, err.Error(), status)
		return
	}

	// TODO: Implement metrics collection and formatting
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("# TODO: Implement Prometheus metrics format"))
}
