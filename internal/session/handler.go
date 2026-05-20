package session

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/lq5657/talkent/internal/role"
)

type Handler struct {
	svc    *Service
	logger *slog.Logger
}

func NewHandler(svc *Service, logger *slog.Logger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/sessions", h.handleCreate)
	mux.HandleFunc("POST /api/sessions/{id}/chat", h.handleChat)
	mux.HandleFunc("GET /api/sessions/{id}/chat/stream", h.handleChatStream)
	mux.HandleFunc("POST /api/sessions/{id}/end", h.handleEnd)
	mux.HandleFunc("GET /api/sessions/{id}", h.handleGet)
}

type createSessionRequest struct {
	RoleDescription string             `json:"role_description"`
	Scenario        string             `json:"scenario"`
	RoleType        role.RoleType      `json:"role_type"`
	Goals           []role.TrainingGoal `json:"goals"`
	Dimensions      []role.Dimension   `json:"dimensions"`
	RoundLimit      int                `json:"round_limit"`
}

type createSessionResponse struct {
	SessionID  string `json:"session_id"`
	Status     string `json:"status"`
	RoundLimit int    `json:"round_limit"`
	CreatedAt  string `json:"created_at"`
}

func (h *Handler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req createSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid request body", "error", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.RoleDescription == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "role_description is required"})
		return
	}

	sess, err := h.svc.CreateSession(r.Context(), CreateSessionRequest{
		RoleDescription: req.RoleDescription,
		Scenario:        req.Scenario,
		RoleType:        req.RoleType,
		Goals:           req.Goals,
		Dimensions:      req.Dimensions,
		RoundLimit:      req.RoundLimit,
	})
	if err != nil {
		h.logger.Error("create session failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create session"})
		return
	}

	writeJSON(w, http.StatusCreated, createSessionResponse{
		SessionID:  sess.ID,
		Status:     sess.Status,
		RoundLimit: sess.RoundLimit,
		CreatedAt:  sess.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

type chatRequest struct {
	Content string `json:"content"`
}

type chatResponse struct {
	Reply                     string    `json:"reply"`
	RoundInfo                 roundInfo `json:"round_info"`
	MemorySource              string    `json:"memory_source"`
	UserMessageCreatedAt      string    `json:"user_message_created_at"`
	AssistantMessageCreatedAt string    `json:"assistant_message_created_at"`
}

type roundInfo struct {
	Current int  `json:"current"`
	Limit   int  `json:"limit"`
	IsLast  bool `json:"is_last"`
}

func (h *Handler) handleChat(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")

	var req chatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid request body", "error", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Content == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "content is required"})
		return
	}

	result, err := h.svc.Chat(r.Context(), sessionID, req.Content)
	if err != nil {
		switch err {
		case ErrSessionNotFound:
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "session not found"})
		case ErrSessionCompleted:
			writeJSON(w, http.StatusConflict, map[string]string{"error": "session already completed"})
		default:
			h.logger.Error("chat failed", "error", err, "session_id", sessionID)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to process chat"})
		}
		return
	}

	writeJSON(w, http.StatusOK, chatResponse{
		Reply: result.Reply,
		RoundInfo: roundInfo{
			Current: result.CurrentRound,
			Limit:   result.RoundLimit,
			IsLast:  result.IsLast,
		},
		MemorySource:              result.MemorySource,
		UserMessageCreatedAt:      result.UserMessageCreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		AssistantMessageCreatedAt: result.AssistantMessageCreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

type endSessionResponse struct {
	SessionID  string `json:"session_id"`
	Status     string `json:"status"`
	FinalRound int    `json:"final_round"`
}

func (h *Handler) handleEnd(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")

	result, err := h.svc.EndSession(r.Context(), sessionID)
	if err != nil {
		switch err {
		case ErrSessionNotFound:
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "session not found"})
		case ErrSessionCompleted:
			writeJSON(w, http.StatusConflict, map[string]string{"error": "session already completed"})
		default:
			h.logger.Error("end session failed", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to end session"})
		}
		return
	}

	writeJSON(w, http.StatusOK, endSessionResponse{
		SessionID:  result.SessionID,
		Status:     result.Status,
		FinalRound: result.FinalRound,
	})
}

type getSessionResponse struct {
	SessionID       string `json:"session_id"`
	Status          string `json:"status"`
	RoleDescription string `json:"role_description"`
	RoundLimit      int    `json:"round_limit"`
	CurrentRound    int    `json:"current_round"`
	MessageCount    int    `json:"message_count"`
	CreatedAt       string `json:"created_at"`
}

func (h *Handler) handleGet(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")

	detail, err := h.svc.GetSessionDetail(r.Context(), sessionID)
	if err != nil {
		if err == ErrSessionNotFound {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "session not found"})
			return
		}
		h.logger.Error("get session failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get session"})
		return
	}

	writeJSON(w, http.StatusOK, getSessionResponse{
		SessionID:       detail.SessionID,
		Status:          detail.Status,
		RoleDescription: detail.RoleDescription,
		RoundLimit:      detail.RoundLimit,
		CurrentRound:    detail.CurrentRound,
		MessageCount:    detail.MessageCount,
		CreatedAt:       detail.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

func (h *Handler) handleChatStream(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	content := r.URL.Query().Get("content")

	if content == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "content is required"})
		return
	}

	ch, err := h.svc.ChatStream(r.Context(), sessionID, content)
	if err != nil {
		switch err {
		case ErrSessionNotFound:
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "session not found"})
		case ErrSessionCompleted:
			writeJSON(w, http.StatusConflict, map[string]string{"error": "session already completed"})
		default:
			h.logger.Error("chat stream failed", "error", err, "session_id", sessionID)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to process chat"})
		}
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "streaming not supported"})
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	for chunk := range ch {
		if chunk.Error != nil {
			fmt.Fprintf(w, "data: {\"error\":%q}\n\n", chunk.Error.Error())
			flusher.Flush()
			return
		}

		if chunk.Done {
			fmt.Fprintf(w, "data: {\"done\":true,\"reply\":%q,\"round_info\":{\"current\":%d,\"limit\":%d,\"is_last\":%v},\"memory_source\":%q,\"user_message_created_at\":%q,\"assistant_message_created_at\":%q}\n\n",
				chunk.Result.Reply,
				chunk.Result.CurrentRound,
				chunk.Result.RoundLimit,
				chunk.Result.IsLast,
				chunk.Result.MemorySource,
				chunk.Result.UserMessageCreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
				chunk.Result.AssistantMessageCreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
			)
			flusher.Flush()
			return
		}

		fmt.Fprintf(w, "data: {\"token\":%q}\n\n", chunk.Content)
		flusher.Flush()
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
