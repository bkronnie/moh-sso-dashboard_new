package feedback

import (
	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/authz"
	"github.com/moh-sso-dashboard/internal/middleware"
)

func RegisterProtectedRoutes(protected *gin.RouterGroup, handler *Handler) {
	fb := protected.Group("/feedback")
	{
		fb.GET("", middleware.RequirePermission(authz.PermissionFeedbackRead), handler.List)
		fb.PATCH("/:id/triage", middleware.RequirePermission(authz.PermissionFeedbackTriage), handler.UpdateTriage)
	}
}
