package repositories

import (
	"context"
	"github.com/google/uuid"
	"github.com/red-velvet-workspace/banco-digital/internal/domain/models"
	"gorm.io/gorm"
)

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(ctx context.Context, notification *models.Notification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}

func (r *NotificationRepository) GetByAccountID(ctx context.Context, accountID uuid.UUID) ([]models.Notification, error) {
	var notifications []models.Notification
	err := r.db.WithContext(ctx).
		Where("account_id = ?", accountID).
		Order("created_at DESC").
		Find(&notifications).Error
	return notifications, err
}

func (r *NotificationRepository) MarkAsRead(ctx context.Context, notificationID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("id = ?", notificationID).
		Update("read", true).Error
}

func (r *NotificationRepository) DeleteOldNotifications(ctx context.Context, accountID uuid.UUID, olderThan string) error {
	return r.db.WithContext(ctx).
		Where("account_id = ? AND created_at < NOW() - INTERVAL ?", accountID, olderThan).
		Delete(&models.Notification{}).Error
}
