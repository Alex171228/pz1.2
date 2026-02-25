package http

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"pz1.2/services/tasks/internal/client/authclient"
	"pz1.2/services/tasks/internal/service"
	"pz1.2/shared/middleware"
)

type Handler struct {
	taskService  *service.TaskService
	authVerifier authclient.AuthVerifier
}

func NewHandler(taskService *service.TaskService, authVerifier authclient.AuthVerifier) *Handler {
	return &Handler{
		taskService:  taskService,
		authVerifier: authVerifier,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/tasks", h.authMiddleware(h.handleCreate))
	mux.HandleFunc("GET /v1/tasks", h.authMiddleware(h.handleGetAll))
	mux.HandleFunc("GET /v1/tasks/{id}", h.authMiddleware(h.handleGetByID))
	mux.HandleFunc("PATCH /v1/tasks/{id}", h.authMiddleware(h.handleUpdate))
	mux.HandleFunc("DELETE /v1/tasks/{id}", h.authMiddleware(h.handleDelete))
}

func (h *Handler) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetRequestID(r.Context())

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Printf("[%s] Missing authorization header", requestID)
			h.respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing authorization header"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			log.Printf("[%s] Invalid authorization format", requestID)
			h.respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid authorization format"})
			return
		}

		token := parts[1]

		verifyResp, err := h.authVerifier.Verify(r.Context(), token)
		if err != nil {
			log.Printf("[%s] Auth service unavailable: %v", requestID, err)
			h.respondJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "auth service unavailable"})
			return
		}

		if !verifyResp.Valid {
			log.Printf("[%s] Invalid token", requestID)
			h.respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
			return
		}

		log.Printf("[%s] Token verified for subject: %s", requestID, verifyResp.Subject)
		next(w, r)
	}
}

func (h *Handler) handleCreate(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	log.Printf("[%s] Creating new task", requestID)

	var req service.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Title == "" {
		h.respondJSON(w, http.StatusBadRequest, map[string]string{"error": "title is required"})
		return
	}

	task := h.taskService.Create(req)
	log.Printf("[%s] Task created: %s", requestID, task.ID)
	h.respondJSON(w, http.StatusCreated, task)
}

func (h *Handler) handleGetAll(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	log.Printf("[%s] Getting all tasks", requestID)

	tasks := h.taskService.GetAll()
	h.respondJSON(w, http.StatusOK, tasks)
}

func (h *Handler) handleGetByID(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	id := r.PathValue("id")
	log.Printf("[%s] Getting task: %s", requestID, id)

	task, err := h.taskService.GetByID(id)
	if err != nil {
		h.respondJSON(w, http.StatusNotFound, map[string]string{"error": "task not found"})
		return
	}

	h.respondJSON(w, http.StatusOK, task)
}

func (h *Handler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	id := r.PathValue("id")
	log.Printf("[%s] Updating task: %s", requestID, id)

	var req service.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	task, err := h.taskService.Update(id, req)
	if err != nil {
		h.respondJSON(w, http.StatusNotFound, map[string]string{"error": "task not found"})
		return
	}

	log.Printf("[%s] Task updated: %s", requestID, id)
	h.respondJSON(w, http.StatusOK, task)
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	id := r.PathValue("id")
	log.Printf("[%s] Deleting task: %s", requestID, id)

	if err := h.taskService.Delete(id); err != nil {
		h.respondJSON(w, http.StatusNotFound, map[string]string{"error": "task not found"})
		return
	}

	log.Printf("[%s] Task deleted: %s", requestID, id)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}
