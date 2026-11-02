package bootstrap

import (
	"database/sql"

	"github.com/moh-sso-dashboard/internal/config"
	storepkg "github.com/moh-sso-dashboard/internal/db/sqlc"
	announcementRepo "github.com/moh-sso-dashboard/internal/features/announcements"
	clientRepo "github.com/moh-sso-dashboard/internal/features/clients"
	documentTemplateRepo "github.com/moh-sso-dashboard/internal/features/document_templates"
	documentRepo "github.com/moh-sso-dashboard/internal/features/documents"
	feedbackfeature "github.com/moh-sso-dashboard/internal/features/feedback"
	rbacRepo "github.com/moh-sso-dashboard/internal/features/rbac"
	sessionRepo "github.com/moh-sso-dashboard/internal/features/sessions"
	storageLocationRepo "github.com/moh-sso-dashboard/internal/features/storage_locations"
	surveillanceRepo "github.com/moh-sso-dashboard/internal/features/surveillance"
	userRepo "github.com/moh-sso-dashboard/internal/features/users"
	kcClientPkg "github.com/moh-sso-dashboard/internal/keycloak"
	logger "github.com/moh-sso-dashboard/internal/log"
	authRepo "github.com/moh-sso-dashboard/internal/repository/auth"
	emailRepo "github.com/moh-sso-dashboard/internal/repository/email"
	metricsRepo "github.com/moh-sso-dashboard/internal/repository/metrics"
	notificationDeliveryRepo "github.com/moh-sso-dashboard/internal/repository/notification_delivery"
	notificationPreferencesRepo "github.com/moh-sso-dashboard/internal/repository/notification_preferences"
	notificationsRepo "github.com/moh-sso-dashboard/internal/repository/notifications"
	processRepo "github.com/moh-sso-dashboard/internal/repository/processes"
)

type repositories struct {
	Auth                    authRepo.AuthRepository
	Clients                 clientRepo.ClientRepository
	Users                   userRepo.UserRepository
	Metrics                 metricsRepo.MetricsRepository
	Notifications           notificationsRepo.NotificationsRepository
	NotificationDelivery    notificationDeliveryRepo.NotificationDeliveryRepository
	NotificationPreferences notificationPreferencesRepo.NotificationPreferencesRepository
	Documents               documentRepo.DocumentRepository
	DocumentFiles           documentRepo.FileRepository
	DocumentTemplates       documentTemplateRepo.DocumentTemplateRepository
	DocumentTemplateColumns documentTemplateRepo.DocumentTemplateColumnRepository
	DocumentTemplateSheets  documentTemplateRepo.DocumentTemplateSheetRepository
	Processes               processRepo.ProcessRepository
	StorageLocations        storageLocationRepo.StorageLocationRepository
	Sessions                sessionRepo.SessionRepository
	Announcements           announcementRepo.AnnouncementRepository
	Email                   emailRepo.EmailRepository
	Surveillance            *surveillanceRepo.Repositories
	RBAC                    rbacRepo.Repository
	Feedback                feedbackfeature.Repository
}

func buildRepositories(
	cfg *config.Config,
	store storepkg.Store,
	adminKC *kcClientPkg.KeyAdminClient,
	webKC *kcClientPkg.Client,
	appLogger *logger.Logger,
) repositories {
	var primaryDB *sql.DB
	if sqlStore, ok := store.(*storepkg.SQLStore); ok && sqlStore != nil {
		primaryDB = sqlStore.DB()
	}

	return repositories{
		Auth:                    authRepo.NewAuthRepository(webKC, adminKC, cfg),
		Clients:                 clientRepo.NewClientRepository(adminKC, cfg, store, *appLogger),
		Users:                   userRepo.NewUserRepository(adminKC, cfg, store, *appLogger),
		Metrics:                 metricsRepo.NewMetricsRepository(cfg, store, *appLogger),
		Notifications:           notificationsRepo.NewNotificationsRepository(store, *appLogger),
		NotificationDelivery:    notificationDeliveryRepo.NewNotificationDeliveryRepository(store, *appLogger),
		NotificationPreferences: notificationPreferencesRepo.NewNotificationPreferencesRepository(store),
		Documents:               documentRepo.NewDocumentRepository(cfg, store, *appLogger),
		DocumentFiles:           documentRepo.NewFileRepository(),
		DocumentTemplates:       documentTemplateRepo.NewDocumentTemplateRepository(store),
		DocumentTemplateColumns: documentTemplateRepo.NewDocumentTemplateColumnRepository(store),
		DocumentTemplateSheets:  documentTemplateRepo.NewDocumentTemplateSheetRepository(store),
		Processes:               processRepo.NewProcessRepository(cfg, store, *appLogger),
		StorageLocations:        storageLocationRepo.NewStorageRepositoryRepository(cfg, store, *appLogger),
		Sessions:                sessionRepo.NewSessionRepository(adminKC, cfg, *appLogger),
		Announcements:           announcementRepo.NewAnnouncementRepository(store, *appLogger),
		Email:                   emailRepo.NewEmailRepository(cfg, store, *appLogger),
		Surveillance:            surveillanceRepo.NewRepositories(store),
		RBAC:                    rbacRepo.NewRepository(primaryDB),
		Feedback:                feedbackfeature.NewHTTPRepository(cfg),
	}
}
