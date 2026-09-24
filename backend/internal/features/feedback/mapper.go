package feedback

func toFeedbackResponse(e FeedbackEntry) FeedbackResponse {
	display := e.SectionTitle
	if display == "" {
		display = e.SectionID
	}

	return FeedbackResponse{
		ID:             e.ID,
		ReportID:       e.ReportID,
		SectionID:      e.SectionID,
		SectionTitle:   e.SectionTitle,
		SectionDisplay: display,
		SubsectionID:   e.SubsectionID,
		ComponentID:    e.ComponentID,
		ComponentTitle: e.ComponentTitle,
		Username:       e.Username,
		Email:          e.Email,

		Institution:      e.Institution,
		ReporterRole:     e.ReporterRole,
		LevelOfOperation: e.LevelOfOperation,
		Phone:            e.Phone,

		FeedbackType:              e.FeedbackType,
		Severity:                  e.Severity,
		Comment:                   e.Comment,
		ExpectedResult:            e.ExpectedResult,
		AffectsDecisionMaking:     e.AffectsDecisionMaking,
		ImpactDescription:         e.ImpactDescription,
		FollowUpNeeded:            e.FollowUpNeeded,
		CorrectiveActionCandidate: e.CorrectiveActionCandidate,
		Recommendation:            e.Recommendation,

		Status:         e.Status,
		AssignedTo:     e.AssignedTo,
		SignOffBlocker: e.SignOffBlocker,

		JiraProject: e.JiraProject,
		JiraEpic:    e.JiraEpic,
		JiraSprint:  e.JiraSprint,
		JiraKey:     e.JiraKey,

		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,

		FiltersSnapshot:      e.FiltersSnapshot,
		HasComponentSnapshot: e.HasComponentSnapshot,
	}
}

func toFeedbackResponses(entries []FeedbackEntry) []FeedbackResponse {
	out := make([]FeedbackResponse, 0, len(entries))
	for _, e := range entries {
		out = append(out, toFeedbackResponse(e))
	}
	return out
}
