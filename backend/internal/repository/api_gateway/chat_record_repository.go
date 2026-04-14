package apigatewayrepo

import (
	"backend/internal/models/entities"
	"strings"

	"gorm.io/gorm"
)

type ChatRecordRepository struct {
	db *gorm.DB
}

func NewChatRecordRepository(db *gorm.DB) *ChatRecordRepository {
	return &ChatRecordRepository{db: db}
}

func (r *ChatRecordRepository) Create(item *entities.ChatRecord) error {
	return r.db.Create(item).Error
}

func (r *ChatRecordRepository) ListByUserID(userID uint, interfaceID uint, keyword string, statusCode int, page int, pageSize int) ([]entities.ChatRecord, int64, error) {
	query := r.db.Model(&entities.ChatRecord{}).Where("user_id = ?", userID)

	if interfaceID > 0 {
		query = query.Where("user_interface_id = ?", interfaceID)
	}
	if statusCode > 0 {
		query = query.Where("status_code = ?", statusCode)
	}
	if trimmed := strings.TrimSpace(keyword); trimmed != "" {
		likeKeyword := "%" + trimmed + "%"
		query = query.Where(
			"request_path ILIKE ? OR model ILIKE ? OR user_input ILIKE ? OR model_output ILIKE ? OR error_message ILIKE ?",
			likeKeyword,
			likeKeyword,
			likeKeyword,
			likeKeyword,
			likeKeyword,
		)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []entities.ChatRecord
	err := query.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error
	return items, total, err
}
