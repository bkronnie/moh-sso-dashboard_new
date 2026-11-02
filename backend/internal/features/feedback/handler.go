package feedback

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/moh-sso-dashboard/internal/http/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// List handles GET /feedback — filters mirror dwh_visualization's
// GET /api/feedback query params (ListFilter, repository.go), renamed to
// this backend's snake_case convention.
func (h *Handler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	var before int64
	if v := c.Query("before"); v != "" {
		before, _ = strconv.ParseInt(v, 10, 64)
	}

	filter := ListFilter{
		ReportID:     c.Query("report_id"),
		ProgramArea:  c.Query("program_area"),
		SectionTitle: c.Query("section_title"),
		Status:       c.Query("status"),
		Severity:     c.Query("severity"),
		FeedbackType: c.Query("type"),
		AssignedTo:   c.Query("assigned_to"),
		Limit:        limit,
		Before:       before,
	}

	entries, hasMore, nextCursor, err := h.service.List(c.Request.Context(), filter)
	if err != nil {
		writeFeedbackError(c, err, "LIST_FAILED", "failed to load feedback from dwh_visualization")
		return
	}

	response.OK(c, http.StatusOK, FeedbackListResponse{
		Entries:    toFeedbackResponses(entries),
		HasMore:    hasMore,
		NextCursor: nextCursor,
	})
}

// UpdateTriage handles PATCH /feedback/:id/triage.
func (h *Handler) UpdateTriage(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Fail(c, http.StatusBadRequest, "INVALID_ID", "invalid feedback id")
		return
	}

	var req triageUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_PAYLOAD", "invalid payload")
		return
	}

	update := TriageUpdate{
		Status:         req.Status,
		AssignedTo:     req.AssignedTo,
		SignOffBlocker: req.SignOffBlocker,
		JiraProject:    req.JiraProject,
		JiraEpic:       req.JiraEpic,
		JiraSprint:     req.JiraSprint,
		JiraKey:        req.JiraKey,
	}

	if err := h.service.UpdateTriage(c.Request.Context(), id, update); err != nil {
		switch {
		case errors.Is(err, ErrInvalidStatus):
			response.Fail(c, http.StatusBadRequest, "INVALID_STATUS", "invalid status")
		default:
			writeFeedbackError(c, err, "TRIAGE_UPDATE_FAILED", "failed to update triage fields")
		}
		return
	}

	response.OK(c, http.StatusOK, triageUpdateResponse{Message: "updated"})
}

// writeFeedbackError maps the two proxy-specific sentinel errors
// (repository.go) to their HTTP shape, falling back to 502 for anything
// else — an unrecognized failure calling dwh_visualization is an upstream
// problem, not a client error.
func writeFeedbackError(c *gin.Context, err error, fallbackCode, fallbackMessage string) {
	switch {
	case errors.Is(err, ErrNotFound):
		response.Fail(c, http.StatusNotFound, "NOT_FOUND", "feedback entry not found")
	case errors.Is(err, ErrNotConfigured):
		response.Fail(c, http.StatusServiceUnavailable, "FEEDBACK_PROXY_NOT_CONFIGURED", "the feedback service is not configured")
	default:
		response.Fail(c, http.StatusBadGateway, fallbackCode, fallbackMessage)
	}
}
