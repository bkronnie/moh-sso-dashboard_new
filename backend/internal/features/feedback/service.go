package feedback

import (
	"context"
	"errors"
	"strings"
)

// ErrInvalidStatus is returned when a triage update's status isn't one of
// the fixed workflow values dwh_visualization itself enforces
// (feedbackValidStatuses, server/feedback_handlers.go). Duplicated here
// (not re-derived from a live call) so a bad status fails fast, without a
// round trip to dwh_visualization — six fixed values, low risk of drift.
var ErrInvalidStatus = errors.New("invalid status")

var validStatuses = map[string]bool{
	"New":                    true,
	"Assigned":               true,
	"Correction in progress": true,
	"Revalidation pending":   true,
	"Resolved":               true,
	"Closed":                 true,
}

const defaultListLimit = 50

type Service interface {
	List(ctx context.Context, filter ListFilter) (entries []FeedbackEntry, hasMore bool, nextCursor int64, err error)
	UpdateTriage(ctx context.Context, id int64, update TriageUpdate) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context, filter ListFilter) ([]FeedbackEntry, bool, int64, error) {
	if s == nil || s.repo == nil {
		return nil, false, 0, errors.New("feedback service is nil")
	}

	if filter.Limit <= 0 || filter.Limit > 200 {
		filter.Limit = defaultListLimit
	}

	return s.repo.List(ctx, filter)
}

func (s *service) UpdateTriage(ctx context.Context, id int64, update TriageUpdate) error {
	if s == nil || s.repo == nil {
		return errors.New("feedback service is nil")
	}
	if id <= 0 {
		return errors.New("invalid feedback id")
	}

	status := strings.TrimSpace(update.Status)
	if !validStatuses[status] {
		return ErrInvalidStatus
	}
	update.Status = status

	if update.SignOffBlocker != "Yes" && update.SignOffBlocker != "No" {
		update.SignOffBlocker = "No"
	}

	return s.repo.UpdateTriage(ctx, id, update)
}
