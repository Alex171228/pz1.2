package http

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"pz1.2/services/auth/internal/service"
	"pz1.2/shared/middleware"
)

type Handler struct {
	authService *service.AuthService
}

func NewHandler(authService *service.AuthService) *Handler {
	return &Handler{authService: authService}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/auth/login", h.handleLogin)
	mux.HandleFunc("GET /v1/auth/verify", h.handleVerify)
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	log.Printf("[%s] Processing login request", requestID)

	var req service.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	resp, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		log.Printf("[%s] Login failed: %v", requestID, err)
		h.respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	log.Printf("[%s] Login successful for user: %s", requestID, req.Username)
	h.respondJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleVerify(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	log.Printf("[%s] Processing verify request", requestID)

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		h.respondJSON(w, http.StatusUnauthorized, service.VerifyResponse{
			Valid: false,
			Error: "missing authorization header",
		})
		return
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		h.respondJSON(w, http.StatusUnauthorized, service.VerifyResponse{
			Valid: false,
			Error: "invalid authorization format",
		})
		return
	}

	token := parts[1]
	resp, err := h.authService.Verify(token)
	if err != nil {
		log.Printf("[%s] Token verification failed: %v", requestID, err)
		h.respondJSON(w, http.StatusUnauthorized, resp)
		return
	}

	log.Printf("[%s] Token verified for subject: %s", requestID, resp.Subject)
	h.respondJSON(w, http.StatusOK, resp)
}

func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
