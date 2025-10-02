package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/b25/services/messaging/internal/models"
	"github.com/b25/services/messaging/internal/service"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

// Handler handles HTTP requests
type Handler struct {
	service *service.MessagingService
	logger  zerolog.Logger
}

// NewHandler creates a new API handler
func NewHandler(svc *service.MessagingService, logger zerolog.Logger) *Handler {
	return &Handler{
		service: svc,
		logger:  logger,
	}
}

// Response helpers

func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{"error": message})
}

// Conversation handlers

func (h *Handler) CreateConversation(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)

	var req models.CreateConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	conv, err := h.service.CreateConversation(r.Context(), userID, &req)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to create conversation")
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusCreated, conv)
}

func (h *Handler) GetConversation(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)
	vars := mux.Vars(r)

	conversationID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid conversation ID")
		return
	}

	conv, err := h.service.GetConversation(r.Context(), conversationID, userID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get conversation")
		h.respondError(w, http.StatusNotFound, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, conv)
}

func (h *Handler) ListConversations(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	convs, err := h.service.ListUserConversations(r.Context(), userID, limit, offset)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to list conversations")
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, convs)
}

func (h *Handler) UpdateConversation(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)
	vars := mux.Vars(r)

	conversationID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid conversation ID")
		return
	}

	var req models.UpdateConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	conv, err := h.service.UpdateConversation(r.Context(), conversationID, userID, &req)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to update conversation")
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, conv)
}

func (h *Handler) AddMember(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)
	vars := mux.Vars(r)

	conversationID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid conversation ID")
		return
	}

	var req struct {
		UserID string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	newMemberID, err := uuid.Parse(req.UserID)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if err := h.service.AddConversationMember(r.Context(), conversationID, userID, newMemberID); err != nil {
		h.logger.Error().Err(err).Msg("Failed to add member")
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)
	vars := mux.Vars(r)

	conversationID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid conversation ID")
		return
	}

	memberID, err := uuid.Parse(vars["userId"])
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if err := h.service.RemoveConversationMember(r.Context(), conversationID, userID, memberID); err != nil {
		h.logger.Error().Err(err).Msg("Failed to remove member")
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Message handlers

func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)
	vars := mux.Vars(r)

	conversationID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid conversation ID")
		return
	}

	var req models.SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	msg, err := h.service.SendMessage(r.Context(), conversationID, userID, &req)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to send message")
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusCreated, msg)
}

func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)
	vars := mux.Vars(r)

	conversationID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid conversation ID")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	messages, err := h.service.GetConversationMessages(r.Context(), conversationID, userID, limit, offset)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get messages")
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, messages)
}

func (h *Handler) EditMessage(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)
	vars := mux.Vars(r)

	messageID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid message ID")
		return
	}

	var req models.EditMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	msg, err := h.service.EditMessage(r.Context(), messageID, userID, req.Content)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to edit message")
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, msg)
}

func (h *Handler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)
	vars := mux.Vars(r)

	messageID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid message ID")
		return
	}

	if err := h.service.DeleteMessage(r.Context(), messageID, userID); err != nil {
		h.logger.Error().Err(err).Msg("Failed to delete message")
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) AddReaction(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)
	vars := mux.Vars(r)

	messageID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid message ID")
		return
	}

	var req models.AddReactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.AddReaction(r.Context(), messageID, userID, req.Emoji); err != nil {
		h.logger.Error().Err(err).Msg("Failed to add reaction")
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) RemoveReaction(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)
	vars := mux.Vars(r)

	messageID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid message ID")
		return
	}

	emoji := vars["emoji"]
	if emoji == "" {
		h.respondError(w, http.StatusBadRequest, "Emoji is required")
		return
	}

	if err := h.service.RemoveReaction(r.Context(), messageID, userID, emoji); err != nil {
		h.logger.Error().Err(err).Msg("Failed to remove reaction")
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)
	vars := mux.Vars(r)

	messageID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid message ID")
		return
	}

	if err := h.service.MarkMessageAsRead(r.Context(), messageID, userID); err != nil {
		h.logger.Error().Err(err).Msg("Failed to mark message as read")
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Search handler

func (h *Handler) SearchMessages(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)

	query := r.URL.Query().Get("q")
	if query == "" {
		h.respondError(w, http.StatusBadRequest, "Query parameter 'q' is required")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	messages, err := h.service.SearchMessages(r.Context(), userID, query, limit, offset)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to search messages")
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, messages)
}

// Presence handlers

func (h *Handler) GetOnlineUsers(w http.ResponseWriter, r *http.Request) {
	users := h.service.GetOnlineUsers()
	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"users": users,
		"count": len(users),
	})
}

// Health check

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	h.respondJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}
